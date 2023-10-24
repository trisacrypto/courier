package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

// New creates a new API client that implements the CourierClient interface.
func New(endpoint string, opts ...ClientOption) (_ CourierClient, err error) {
	if endpoint == "" {
		return nil, ErrEndpointRequired
	}

	// Create a client with the parsed endpoint.
	c := &APIv1{}
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
	return c, nil
}

// APIv1 implements the CourierClient interface.
type APIv1 struct {
	url    *url.URL
	client *http.Client
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
