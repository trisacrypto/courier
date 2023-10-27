package gcloud

import "github.com/trisacrypto/courier/pkg/secrets"

// StoreOption allows us to configure the store when it is created.
type StoreOption func(s *Store) error

func WithClient(client secrets.SecretManagerClient) StoreOption {
	return func(s *Store) error {
		s.client = client
		return nil
	}
}
