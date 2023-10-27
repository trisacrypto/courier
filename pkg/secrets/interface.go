package secrets

import (
	"context"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go"
)

// SecretManagerClient describes a high level interface for secret manager clients to
// enable mocking.
type SecretManagerClient interface {
	GetLatestVersion(ctx context.Context, name string) ([]byte, error)
	CreateSecret(ctx context.Context, name string) error
	AddSecretVersion(ctx context.Context, name string, payload []byte) error
	DeleteSecret(ctx context.Context, name string) error
}

// gRPCSecretClient describes a lower level interface in order to mock the google secret
// manager client.
type GRPCSecretClient interface {
	CreateSecret(context.Context, *secretmanagerpb.CreateSecretRequest, ...gax.CallOption) (*secretmanagerpb.Secret, error)
	GetSecretVersion(context.Context, *secretmanagerpb.GetSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	AddSecretVersion(context.Context, *secretmanagerpb.AddSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	AccessSecretVersion(context.Context, *secretmanagerpb.AccessSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	DeleteSecret(context.Context, *secretmanagerpb.DeleteSecretRequest, ...gax.CallOption) error
}
