package secrets

import (
	"context"
)

// secretManagerClient describes a high level interface for secret manager clients to
// enable mocking.
type secretManagerClient interface {
	GetLatestVersion(ctx context.Context, name string) ([]byte, error)
	CreateSecret(ctx context.Context, name string) error
	AddSecretVersion(ctx context.Context, name string, payload []byte) error
	DeleteSecret(ctx context.Context, name string) error
}
