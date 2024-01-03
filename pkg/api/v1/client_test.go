package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/courier/pkg/api/v1"
)

func TestStoreCertificate(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/certs/1234", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	// Create a client to test the client method
	client, err := api.New(ts.URL)
	require.NoError(t, err, "could not create client")

	// Create a new certificate store request
	req := &api.StoreCertificateRequest{
		ID:                "1234",
		Base64Certificate: "base64-encoded-certificate",
	}
	err = client.StoreCertificate(context.Background(), req)
	require.NoError(t, err, "could not execute certificate store request")

	// Should error if there is no ID in the request
	req.ID = ""
	err = client.StoreCertificate(context.Background(), req)
	require.ErrorIs(t, err, api.ErrIDRequired, "client should error if no ID is provided")
}

func TestStoreCertificatePassword(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/certs/1234/pkcs12password", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	// Create a client to test the client method
	client, err := api.New(ts.URL)
	require.NoError(t, err, "could not create client")

	// Create a new password store request
	req := &api.StorePasswordRequest{
		ID:       "1234",
		Password: "hunter2",
	}
	err = client.StoreCertificatePassword(context.Background(), req)
	require.NoError(t, err, "could not execute password store request")

	// Should error if there is no ID in the request
	req.ID = ""
	err = client.StoreCertificatePassword(context.Background(), req)
	require.ErrorIs(t, err, api.ErrIDRequired, "client should error if no ID is provided")
}

func TestRetriesWithBackoff(t *testing.T) {
	// Create a test server
	var attempts uint32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&attempts, 1)
		http.Error(w, http.StatusText(http.StatusTooEarly), http.StatusTooEarly)
	}))
	defer ts.Close()

	// Create a client to test the client method
	client, err := api.New(ts.URL, api.WithRetries(10), api.WithBackoff(func() backoff.BackOff {
		return backoff.NewConstantBackOff(100 * time.Millisecond)
	}))
	require.NoError(t, err, "could not create client")

	rawClient, ok := client.(*api.APIv1)
	require.True(t, ok, "expected client to be an APIv1 client")

	req, err := rawClient.NewRequest(context.Background(), http.MethodGet, "/", nil, nil)
	require.NoError(t, err, "could not create request")

	start := time.Now()
	_, err = rawClient.Do(req, nil, true)
	require.Error(t, err, "expected an error to be returned")
	require.Equal(t, uint32(11), attempts, "expected 10 retry attempts")
	require.Greater(t, time.Since(start), 950*time.Millisecond, "expected backoff delay")
}
