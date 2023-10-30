package courier_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/trisacrypto/courier/pkg/api/v1"
	"github.com/trisacrypto/courier/pkg/store"
	"github.com/trisacrypto/trisa/pkg/trust"
)

func (s *courierTestSuite) TestStoreCertificate() {
	require := s.Require()

	// Load the cert fixture
	sz, err := trust.NewSerializer(true, "supersecretsquirrel")
	require.NoError(err, "could not create serializer")
	provider, err := sz.ReadFile("testdata/cert.zip")
	require.NoError(err, "could not read cert fixture")
	decrypted, err := provider.Encode()
	require.NoError(err, "could not read cert fixture")

	// Encrypt the data for the request
	encrypted, err := provider.Encrypt("supersecretsquirrel")
	require.NoError(err, "could not encrypt cert fixture")
	cert64 := base64.StdEncoding.EncodeToString(encrypted)

	s.Run("HappyPath", func() {
		req := &api.StoreCertificateRequest{
			ID:                "certID",
			Base64Certificate: cert64,
		}

		// Configure the store mock to return the password
		s.store.OnGetPassword = func(ctx context.Context, name string) ([]byte, error) {
			require.Equal(req.ID, name, "wrong cert name passed to get password")
			return []byte("supersecretsquirrel"), nil
		}

		// Configure the store mock to return success on update
		s.store.OnUpdateCertificate = func(ctx context.Context, name string, cert []byte) error {
			require.Equal(req.ID, name, "wrong cert name passed to update cert")
			require.Equal(decrypted, cert, "wrong cert data passed to update cert")
			return nil
		}
		defer s.store.Reset()

		// Make a request to the endpoint
		err := s.client.StoreCertificate(context.Background(), req)
		require.NoError(err, "could not store certificate")
	})

	s.Run("NoDecrypt", func() {
		req := &api.StoreCertificateRequest{
			ID:                "certID",
			Base64Certificate: cert64,
			NoDecrypt:         true,
		}

		// Configure the store mock to return success on update
		s.store.OnUpdateCertificate = func(ctx context.Context, name string, cert []byte) error {
			require.Equal(req.ID, name, "wrong cert name passed to update cert")
			require.Equal(encrypted, cert, "wrong cert data passed to update cert")
			return nil
		}
		defer s.store.Reset()

		// Make a request to the endpoint
		err := s.client.StoreCertificate(context.Background(), req)
		require.NoError(err, "could not store certificate")
	})

	s.Run("MissingCertificate", func() {
		req := &api.StoreCertificateRequest{
			ID: "certID",
		}
		err := s.client.StoreCertificate(context.Background(), req)
		s.CheckHTTPStatus(err, http.StatusBadRequest, "wrong error code for missing certificate")
	})

	s.Run("BadBase64", func() {
		req := &api.StoreCertificateRequest{
			ID:                "certID",
			Base64Certificate: "badbase64",
		}
		err := s.client.StoreCertificate(context.Background(), req)
		s.CheckHTTPStatus(err, http.StatusBadRequest, "wrong error code for bad base64")
	})

	s.Run("MissingPassword", func() {
		req := &api.StoreCertificateRequest{
			ID:                "certID",
			Base64Certificate: cert64,
		}

		// Configure the store mock to return not found on get password
		s.store.OnGetPassword = func(ctx context.Context, name string) ([]byte, error) {
			return nil, store.ErrNotFound
		}
		defer s.store.Reset()

		err := s.client.StoreCertificate(context.Background(), req)
		s.CheckHTTPStatus(err, http.StatusNotFound, "wrong error code for missing password")
	})

	s.Run("WrongCertificate", func() {
		req := &api.StoreCertificateRequest{
			ID:                "certID",
			Base64Certificate: cert64,
		}

		// Configure the store mock to return a password that won't work
		s.store.OnGetPassword = func(ctx context.Context, name string) ([]byte, error) {
			return []byte("wrongpassword"), nil
		}
		defer s.store.Reset()

		err := s.client.StoreCertificate(context.Background(), req)
		s.CheckHTTPStatus(err, http.StatusConflict, "wrong error code for wrong password")
	})

	s.Run("StoreError", func() {
		req := &api.StoreCertificateRequest{
			ID:                "certID",
			Base64Certificate: cert64,
		}

		// Configure the store mock to return an error
		s.store.OnGetPassword = func(ctx context.Context, name string) ([]byte, error) {
			return nil, errors.New("internal store error")
		}
		defer s.store.Reset()

		err := s.client.StoreCertificate(context.Background(), req)
		s.CheckHTTPStatus(err, http.StatusInternalServerError, "wrong error code for store error")
	})
}

func (s *courierTestSuite) TestStoreCertificatePassword() {
	require := s.Require()

	s.Run("HappyPath", func() {
		// Configure the store mock to return a successful response
		req := &api.StorePasswordRequest{
			ID:       "certID",
			Password: "password",
		}
		s.store.OnUpdatePassword = func(ctx context.Context, name string, password []byte) error {
			require.Equal(req.ID, name, "wrong password name passed to store")
			require.Equal([]byte(req.Password), password, "wrong password passed to store")
			return nil
		}
		defer s.store.Reset()

		// Make a request to the endpoint
		err := s.client.StoreCertificatePassword(context.Background(), req)
		require.NoError(err, "could not store certificate password")
	})

	s.Run("MissingPassword", func() {
		req := &api.StorePasswordRequest{
			ID: "certID",
		}
		err := s.client.StoreCertificatePassword(context.Background(), req)
		s.CheckHTTPStatus(err, http.StatusBadRequest, "wrong error code for missing password")
	})

	s.Run("StoreError", func() {
		s.store.OnUpdatePassword = func(ctx context.Context, name string, password []byte) error {
			return errors.New("internal store error")
		}
		defer s.store.Reset()

		req := &api.StorePasswordRequest{
			ID:       "certID",
			Password: "password",
		}
		err := s.client.StoreCertificatePassword(context.Background(), req)
		s.CheckHTTPStatus(err, http.StatusInternalServerError, "wrong error code for store error")
	})
}
