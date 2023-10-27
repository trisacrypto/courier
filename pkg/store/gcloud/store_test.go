package gcloud_test

import (
	"context"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go"
	"github.com/stretchr/testify/suite"
	"github.com/trisacrypto/courier/pkg/config"
	"github.com/trisacrypto/courier/pkg/secrets"
	"github.com/trisacrypto/courier/pkg/secrets/mock"
	"github.com/trisacrypto/courier/pkg/store"
	"github.com/trisacrypto/courier/pkg/store/gcloud"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gcloudStoreTestSuite struct {
	suite.Suite
	store *gcloud.Store
	conf  config.SecretsConfig
	sm    *mock.SecretManager
}

func (s *gcloudStoreTestSuite) SetupSuite() {
	// Open the storage backend using a mock secrets client
	var err error
	s.sm = mock.New()
	s.conf = config.SecretsConfig{
		Enabled:     true,
		Credentials: "creds.json",
		Project:     "project",
	}
	client, err := secrets.NewClient(s.conf, secrets.WithGRPCClient(s.sm))
	s.NoError(err, "could not create mock secrets client")
	s.store, err = gcloud.Open(s.conf, gcloud.WithClient(client))
	s.NoError(err, "could not open gcloud storage backend")
}

func (s *gcloudStoreTestSuite) TearDownSuite() {
	// Close the storage backend
	s.NoError(s.store.Close(), "could not close gcloud storage backend")
}

func TestGCloudStore(t *testing.T) {
	suite.Run(t, new(gcloudStoreTestSuite))
}

func (s *gcloudStoreTestSuite) TestGetPassword() {
	require := s.Require()
	ctx := context.Background()

	s.Run("HappyPath", func() {
		s.sm.OnAccessSecretVersion = func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			return &secretmanagerpb.AccessSecretVersionResponse{
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte("password"),
				},
			}, nil
		}
		defer s.sm.Reset()
		password, err := s.store.GetPassword(ctx, "does-exist")
		require.NoError(err, "should be able to get a password")
		require.Equal([]byte("password"), password, "wrong password returned")
	})

	s.Run("NotFound", func() {
		s.sm.OnAccessSecretVersion = func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		defer s.sm.Reset()
		_, err := s.store.GetPassword(ctx, "does-not-exist")
		require.ErrorIs(err, store.ErrNotFound, "should return error if password does not exist")
	})

	s.Run("Error", func() {
		statusErr := status.Error(codes.Internal, "internal error")
		s.sm.OnAccessSecretVersion = func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			return nil, statusErr
		}
		defer s.sm.Reset()
		_, err := s.store.GetPassword(ctx, "does-not-exist")
		require.EqualError(err, statusErr.Error(), "should return error if there was a gRPC error")
	})
}

func (s *gcloudStoreTestSuite) TestUpdatePassword() {
	requre := s.Require()
	ctx := context.Background()

	s.Run("HappyPath", func() {
		s.sm.OnCreateSecret = func(ctx context.Context, req *secretmanagerpb.CreateSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error) {
			return &secretmanagerpb.Secret{}, nil
		}
		s.sm.OnAddSecretVersion = func(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
			return &secretmanagerpb.SecretVersion{}, nil
		}
		defer s.sm.Reset()
		err := s.store.UpdatePassword(ctx, "password_id", []byte("password"))
		requre.NoError(err, "should be able to create a password")
	})

	s.Run("Error", func() {
		statusErr := status.Error(codes.Internal, "internal error")
		s.sm.OnCreateSecret = func(ctx context.Context, req *secretmanagerpb.CreateSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error) {
			return nil, statusErr
		}
		defer s.sm.Reset()
		err := s.store.UpdatePassword(ctx, "password_id", []byte("password"))
		requre.EqualError(err, statusErr.Error(), "should return error if there was a gRPC error")
	})
}

func (s *gcloudStoreTestSuite) TestGetCertificate() {
	require := s.Require()
	ctx := context.Background()

	s.Run("HappyPath", func() {
		s.sm.OnAccessSecretVersion = func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			return &secretmanagerpb.AccessSecretVersionResponse{
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte("cert"),
				},
			}, nil
		}
		defer s.sm.Reset()
		cert, err := s.store.GetCertificate(ctx, "does-exist")
		require.NoError(err, "should be able to get a password")
		require.Equal([]byte("cert"), cert, "wrong cert returned")
	})

	s.Run("NotFound", func() {
		s.sm.OnAccessSecretVersion = func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		defer s.sm.Reset()
		_, err := s.store.GetCertificate(ctx, "does-not-exist")
		require.ErrorIs(err, store.ErrNotFound, "should return error if cert does not exist")
	})

	s.Run("Error", func() {
		statusErr := status.Error(codes.Internal, "internal error")
		s.sm.OnAccessSecretVersion = func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			return nil, statusErr
		}
		defer s.sm.Reset()
		_, err := s.store.GetCertificate(ctx, "does-not-exist")
		require.EqualError(err, statusErr.Error(), "should return error if there was a gRPC error")
	})
}

func (s *gcloudStoreTestSuite) TestUpdateCertificate() {
	requre := s.Require()
	ctx := context.Background()

	s.Run("HappyPath", func() {
		s.sm.OnCreateSecret = func(ctx context.Context, req *secretmanagerpb.CreateSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error) {
			return &secretmanagerpb.Secret{}, nil
		}
		s.sm.OnAddSecretVersion = func(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
			return &secretmanagerpb.SecretVersion{}, nil
		}
		defer s.sm.Reset()
		err := s.store.UpdateCertificate(ctx, "cert_id", []byte("cert"))
		requre.NoError(err, "should be able to create a certificate")
	})

	s.Run("Error", func() {
		statusErr := status.Error(codes.Internal, "internal error")
		s.sm.OnCreateSecret = func(ctx context.Context, req *secretmanagerpb.CreateSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error) {
			return nil, statusErr
		}
		defer s.sm.Reset()
		err := s.store.UpdateCertificate(ctx, "cert_id", []byte("cert"))
		requre.EqualError(err, statusErr.Error(), "should return error if there was a gRPC error")
	})
}
