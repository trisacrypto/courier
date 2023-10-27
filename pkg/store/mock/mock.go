package mock

import "github.com/trisacrypto/courier/pkg/store"

// New returns a new mock store. The On* functions can be used to configure the mock
// behavior directly. Functions that are not configured will return an error.
func New() (s *Store) {
	s = &Store{}
	s.Reset()
	return s
}

// Reset resets the state of the mock so all functions return an error.
func (s *Store) Reset() {
	s.OnGetPassword = func(name string) ([]byte, error) {
		return nil, ErrNotConfigured
	}

	s.OnUpdatePassword = func(name string, password []byte) error {
		return ErrNotConfigured
	}

	s.OnGetCertificate = func(name string) ([]byte, error) {
		return nil, ErrNotConfigured
	}

	s.OnUpdateCertificate = func(name string, cert []byte) error {
		return ErrNotConfigured
	}
}

// Store implements the store.Store interface for mocking the store in tests.
type Store struct {
	OnGetPassword       func(name string) ([]byte, error)
	OnUpdatePassword    func(name string, password []byte) error
	OnGetCertificate    func(name string) ([]byte, error)
	OnUpdateCertificate func(name string, cert []byte) error
}

var _ store.Store = &Store{}

func (s *Store) Close() error {
	return nil
}

func (s *Store) GetPassword(name string) ([]byte, error) {
	return s.OnGetPassword(name)
}

func (s *Store) UpdatePassword(name string, password []byte) error {
	return s.OnUpdatePassword(name, password)
}

func (s *Store) GetCertificate(name string) ([]byte, error) {
	return s.OnGetCertificate(name)
}

func (s *Store) UpdateCertificate(name string, cert []byte) error {
	return s.OnUpdateCertificate(name, cert)
}
