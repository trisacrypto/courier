package mock

import (
	"context"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go"
	"github.com/trisacrypto/courier/pkg/secrets"
)

// New returns a new secrets client mock. The On* functions can be used to configure
// the mock behavior directly. Functions that are not configured will return an error.
func New() (s *SecretManager) {
	s = &SecretManager{}
	s.Reset()
	return s
}

// Reset resets the state of the mock so all functions return an error.
func (s *SecretManager) Reset() {
	s.OnCreateSecret = func(context.Context, *secretmanagerpb.CreateSecretRequest, ...gax.CallOption) (*secretmanagerpb.Secret, error) {
		return nil, ErrNotConfigured
	}
	s.OnGetSecretVersion = func(context.Context, *secretmanagerpb.GetSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
		return nil, ErrNotConfigured
	}
	s.OnAddSecretVersion = func(context.Context, *secretmanagerpb.AddSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
		return nil, ErrNotConfigured
	}
	s.OnAccessSecretVersion = func(context.Context, *secretmanagerpb.AccessSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
		return nil, ErrNotConfigured
	}
	s.OnDeleteSecret = func(context.Context, *secretmanagerpb.DeleteSecretRequest, ...gax.CallOption) error {
		return ErrNotConfigured
	}
}

type SecretManager struct {
	OnCreateSecret        func(context.Context, *secretmanagerpb.CreateSecretRequest, ...gax.CallOption) (*secretmanagerpb.Secret, error)
	OnGetSecretVersion    func(context.Context, *secretmanagerpb.GetSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	OnAddSecretVersion    func(context.Context, *secretmanagerpb.AddSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	OnAccessSecretVersion func(context.Context, *secretmanagerpb.AccessSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	OnDeleteSecret        func(context.Context, *secretmanagerpb.DeleteSecretRequest, ...gax.CallOption) error
}

var _ secrets.GRPCSecretClient = &SecretManager{}

func (s *SecretManager) CreateSecret(ctx context.Context, req *secretmanagerpb.CreateSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error) {
	return s.OnCreateSecret(ctx, req, opts...)
}

func (s *SecretManager) GetSecretVersion(ctx context.Context, req *secretmanagerpb.GetSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	return s.OnGetSecretVersion(ctx, req, opts...)
}

func (s *SecretManager) AddSecretVersion(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	return s.OnAddSecretVersion(ctx, req, opts...)
}

func (s *SecretManager) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return s.OnAccessSecretVersion(ctx, req, opts...)
}

func (s *SecretManager) DeleteSecret(ctx context.Context, req *secretmanagerpb.DeleteSecretRequest, opts ...gax.CallOption) error {
	return s.OnDeleteSecret(ctx, req, opts...)
}
