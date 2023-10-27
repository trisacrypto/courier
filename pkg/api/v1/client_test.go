package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/courier/pkg/api/v1"
)

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

	// Create a new register request
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
