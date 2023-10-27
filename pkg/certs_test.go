package courier_test

import (
	"context"
	"errors"
	"net/http"

	"github.com/trisacrypto/courier/pkg/api/v1"
)

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
