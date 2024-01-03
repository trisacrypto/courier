package api

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// ClientOption allows the API client to be configured when it is created.
type ClientOption func(c *APIv1) error

// BackoffFactory creates a new backoff delay for a specific request.
type BackoffFactory func() backoff.BackOff

// WithBackoff allows the user to create a client that retries requests with a fixed or
// exponential backoff to allow the remote service time to recover. By default, the
// courier client uses exponential backoff and three retries.
func WithBackoff(bf BackoffFactory) ClientOption {
	return func(c *APIv1) error {
		c.backoff = bf
		return nil
	}
}

// WithZeroBackoff creates a client that retries immediately without delay.
func WithZeroBackoff() ClientOption {
	return func(c *APIv1) error {
		c.backoff = func() backoff.BackOff {
			return &backoff.ZeroBackOff{}
		}
		return nil
	}
}

// WithRetries allows the user to create a client that retries requests for the
// specified number of attempts. Set to zero to only send one request with no retries.
// The default number of retry attempts is 3.
func WithRetries(attempts int) ClientOption {
	return func(c *APIv1) error {
		if attempts < 0 {
			return ErrInvalidRetries
		}

		c.retries = attempts
		return nil
	}
}

// WithTLSConfig allows the user to specify a custom tls configuration for the client.
func WithTLSConfig(conf *tls.Config) ClientOption {
	return func(c *APIv1) error {
		if c.client != nil {
			c.client.Transport = &http.Transport{
				TLSClientConfig: conf,
			}
		} else {
			c.client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: conf,
				},
				CheckRedirect: nil,
				Timeout:       30 * time.Second,
			}
		}
		return nil
	}
}
