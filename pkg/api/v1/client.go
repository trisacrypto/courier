package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const DefaultRetries = 3

func DefaultBackoff() BackoffFactory {
	return func() backoff.BackOff {
		return backoff.NewExponentialBackOff()
	}
}

// New creates a new API client that implements the CourierClient interface.
func New(endpoint string, opts ...ClientOption) (_ CourierClient, err error) {
	if endpoint == "" {
		return nil, ErrEndpointRequired
	}

	// Create a client with the parsed endpoint.
	c := &APIv1{retries: -1}
	if c.url, err = url.Parse(endpoint); err != nil {
		return nil, err
	}

	// Apply options
	for _, opt := range opts {
		if err = opt(c); err != nil {
			return nil, err
		}
	}

	// If a client hasn't been specified, create the default client.
	if c.client == nil {
		c.client = &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Timeout:       30 * time.Second,
		}
	}

	// If backoff hasn't been specified add the default backoff factory
	if c.backoff == nil {
		c.backoff = DefaultBackoff()
	}

	// If retries haven't been specified add the default number of retries
	if c.retries < 0 {
		c.retries = DefaultRetries
	}

	return c, nil
}

// APIv1 implements the CourierClient interface.
type APIv1 struct {
	url     *url.URL
	client  *http.Client
	backoff BackoffFactory
	retries int
}

var _ CourierClient = &APIv1{}

//===========================================================================
// Client Methods
//===========================================================================

// Status returns the status of the courier service.
func (c *APIv1) Status(ctx context.Context) (out *StatusReply, err error) {
	// Create the HTTP request
	var req *http.Request
	if req, err = c.NewRequest(ctx, http.MethodGet, "/v1/status", nil, nil); err != nil {
		return nil, err
	}

	// Do the request
	var rep *http.Response
	if rep, err = c.client.Do(req); err != nil {
		return nil, err
	}
	defer rep.Body.Close()

	// Catch status errors
	if rep.StatusCode != http.StatusOK && rep.StatusCode != http.StatusServiceUnavailable {
		return nil, NewStatusError(rep.StatusCode, rep.Status)
	}

	// Decode the response
	out = &StatusReply{}
	if err = json.NewDecoder(rep.Body).Decode(out); err != nil {
		return nil, err
	}
	return out, nil
}

// StoreCertificate stores the certificate in the request.
func (c *APIv1) StoreCertificate(ctx context.Context, in *StoreCertificateRequest) (err error) {
	if in.ID == "" {
		return ErrIDRequired
	}

	path := fmt.Sprintf("/v1/certs/%s", in.ID)

	// Create the HTTP request
	var req *http.Request
	if req, err = c.NewRequest(ctx, http.MethodPost, path, in, nil); err != nil {
		return err
	}

	// Do the request
	if _, err = c.Do(req, nil, true); err != nil {
		return err
	}
	return nil
}

// StoreCertificatePassword stores a password for an encrypted certificate.
func (c *APIv1) StoreCertificatePassword(ctx context.Context, in *StorePasswordRequest) (err error) {
	if in.ID == "" {
		return ErrIDRequired
	}

	path := fmt.Sprintf("/v1/certs/%s/pkcs12password", in.ID)

	// Create the HTTP request
	var req *http.Request
	if req, err = c.NewRequest(ctx, http.MethodPost, path, in, nil); err != nil {
		return err
	}

	// Do the request
	if _, err = c.Do(req, nil, true); err != nil {
		return err
	}
	return nil
}

//===========================================================================
// Client Helpers
//===========================================================================

const (
	userAgent    = "Courier API Client/v1"
	accept       = "application/json"
	acceptLang   = "en-US,en"
	acceptEncode = "gzip, deflate, br"
	contentType  = "application/json; charset=utf-8"
)

// NewRequest creates an http.Request with the specified context and method, resolving
// the path to the root endpoint of the API (e.g. /v1) and serializes the data to JSON.
func (c *APIv1) NewRequest(ctx context.Context, method, path string, data interface{}, params *url.Values) (req *http.Request, err error) {
	// Resolve the URL reference from the path
	endpoint := c.url.ResolveReference(&url.URL{Path: path})
	if params != nil && len(*params) > 0 {
		endpoint.RawQuery = params.Encode()
	}

	var body io.ReadWriter
	switch {
	case data == nil:
		body = nil
	default:
		body = &bytes.Buffer{}
		if err = json.NewEncoder(body).Encode(data); err != nil {
			return nil, err
		}
	}

	// Create the http request
	if req, err = http.NewRequestWithContext(ctx, method, endpoint.String(), body); err != nil {
		return nil, err
	}

	// Set the headers on the request
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept", accept)
	req.Header.Add("Accept-Language", acceptLang)
	req.Header.Add("Accept-Encoding", acceptEncode)
	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// Do executes an http request against the server, performs error checking, and
// deserializes response data into the specified struct. This function also manages
// retries using a backoff strategy.
func (s *APIv1) Do(req *http.Request, data interface{}, checkStatus bool) (rep *http.Response, err error) {
	attempts := 0
	start := time.Now()
	ctx := req.Context()
	delay := s.backoff()
	errs := make([]error, 0, s.retries+1)

	for attempts <= s.retries {
		attempts++
		if rep, err = s.do(req, data, checkStatus); err == nil {
			// Success!
			return rep, nil
		}

		// Failure! Retry as needed.
		errs = append(errs, err)

		// Compute the backoff delay before the next request
		dur := delay.NextBackOff()
		if dur == backoff.Stop {
			// Stop indicates no more retries should be allowed.
			return rep, JoinStatusErrors(attempts, time.Since(start), errs...)
		}

		// Wait for backoff delay or until context is canceled
		wait := time.After(dur)
		select {
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			return rep, JoinStatusErrors(attempts, time.Since(start), errs...)
		case <-wait:
			continue
		}
	}

	return rep, JoinStatusErrors(attempts, time.Since(start), errs...)
}

func (s *APIv1) do(req *http.Request, data interface{}, checkStatus bool) (rep *http.Response, err error) {
	if rep, err = s.client.Do(req); err != nil {
		return rep, err
	}
	defer rep.Body.Close()

	// Detects http status errors if they've occurred
	if checkStatus {
		if rep.StatusCode < 200 || rep.StatusCode >= 300 {
			// Attempt to read the error response from the generic reply
			var reply Reply
			if err = json.NewDecoder(rep.Body).Decode(&reply); err == nil {
				if reply.Error != "" {
					return rep, NewStatusError(rep.StatusCode, reply.Error)
				}
			}
			return rep, NewStatusError(rep.StatusCode, rep.Status)
		}
	}

	// Deserializes the JSON data from the body
	if data != nil && rep.StatusCode >= 200 && rep.StatusCode < 300 && rep.StatusCode != http.StatusNoContent {
		// Checks the content type to ensure data deserialization is possible
		if ct := rep.Header.Get("Content-Type"); ct != contentType {
			return rep, fmt.Errorf("unexpected content type: %q", ct)
		}

		if err = json.NewDecoder(rep.Body).Decode(data); err != nil {
			return nil, fmt.Errorf("could not deserialize response data: %s", err)
		}
	}
	return rep, nil
}
