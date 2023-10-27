package mock

import (
	"context"

	"github.com/trisacrypto/courier/pkg/store"
)

// New returns a new mock store. The On* functions can be used to configure the mock
// behavior directly. Functions that are not configured will return an error.
func New() (s *Store) {
	s = &Store{}
	s.Reset()
	return s
}

// Reset resets the state of the mock so all functions return an error.
func (s *Store) Reset() {
	s.OnGetPassword = func(ctx context.Context, name string) ([]byte, error) {
		return nil, ErrNotConfigured
	}

	s.OnUpdatePassword = func(ctx context.Context, name string, password []byte) error {
		return ErrNotConfigured
	}

	s.OnGetCertificate = func(ctx context.Context, name string) ([]byte, error) {
		return nil, ErrNotConfigured
	}

	s.OnUpdateCertificate = func(ctx context.Context, name string, cert []byte) error {
		return ErrNotConfigured
	}
}

// Store implements the store.Store interface for mocking the store in tests.
type Store struct {
	OnGetPassword       func(ctx context.Context, name string) ([]byte, error)
	OnUpdatePassword    func(ctx context.Context, name string, password []byte) error
	OnGetCertificate    func(ctx context.Context, name string) ([]byte, error)
	OnUpdateCertificate func(ctx context.Context, name string, cert []byte) error
}

var _ store.Store = &Store{}

func (s *Store) Close() error {
	return nil
}

func (s *Store) GetPassword(ctx context.Context, name string) ([]byte, error) {
	return s.OnGetPassword(ctx, name)
}

func (s *Store) UpdatePassword(ctx context.Context, name string, password []byte) error {
	return s.OnUpdatePassword(ctx, name, password)
}

func (s *Store) GetCertificate(ctx context.Context, name string) ([]byte, error) {
	return s.OnGetCertificate(ctx, name)
}

func (s *Store) UpdateCertificate(ctx context.Context, name string, cert []byte) error {
	return s.OnUpdateCertificate(ctx, name, cert)
}
