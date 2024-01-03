package gcloud

import (
	"context"
	"errors"

	"github.com/trisacrypto/courier/pkg/config"
	"github.com/trisacrypto/courier/pkg/secrets"
	"github.com/trisacrypto/courier/pkg/store"
)

// Open the google cloud storage backend.
func Open(conf config.GCPSecretsConfig, opts ...StoreOption) (store *Store, err error) {
	store = &Store{}

	// Apply provided options
	for _, opt := range opts {
		if err = opt(store); err != nil {
			return nil, err
		}
	}

	if store.client == nil {
		if store.client, err = secrets.NewClient(conf); err != nil {
			return nil, err
		}
	}

	return store, nil
}

// Store implements the store.Store interface for google cloud storage using secret
// manager
type Store struct {
	client secrets.SecretManagerClient
}

var _ store.Store = &Store{}

// Close the google cloud storage backend.
func (s *Store) Close() error {
	return nil
}

//===========================================================================
// Password Methods
//===========================================================================

// GetPassword retrieves a password by id from the google cloud storage backend.
func (s *Store) GetPassword(ctx context.Context, id string) (password []byte, err error) {
	if password, err = s.client.GetLatestVersion(ctx, s.fullName(store.PasswordPrefix, id)); err != nil {
		if errors.Is(err, secrets.ErrSecretNotFound) {
			return nil, store.ErrNotFound
		}

		return nil, err
	}

	return password, nil
}

// UpdatePassword updates a password by id in the google cloud storage backend.
func (s *Store) UpdatePassword(ctx context.Context, id string, password []byte) (err error) {
	// Ensure the secret exists, this assumes that an error is not returned if the
	// secret already exists.
	if err = s.client.CreateSecret(ctx, s.fullName(store.PasswordPrefix, id)); err != nil {
		return err
	}

	return s.client.AddSecretVersion(ctx, s.fullName(store.PasswordPrefix, id), password)
}

//===========================================================================
// Certificate Methods
//===========================================================================

// GetCertificate retrieves a certificate by id from the google cloud storage backend.
func (s *Store) GetCertificate(ctx context.Context, id string) (cert []byte, err error) {
	if cert, err = s.client.GetLatestVersion(ctx, s.fullName(store.CertificatePrefix, id)); err != nil {
		if errors.Is(err, secrets.ErrSecretNotFound) {
			return nil, store.ErrNotFound
		}

		return nil, err
	}

	return cert, nil
}

// UpdateCertificate updates a certificate by id in the google cloud storage backend.
func (s *Store) UpdateCertificate(ctx context.Context, id string, cert []byte) (err error) {
	// Ensure the secret exists, this assumes that an error is not returned if the
	// secret already exists.
	if err = s.client.CreateSecret(ctx, s.fullName(store.CertificatePrefix, id)); err != nil {
		return err
	}

	return s.client.AddSecretVersion(ctx, s.fullName(store.CertificatePrefix, id), cert)
}

//===========================================================================
// Helper methods
//===========================================================================

// fullName returns the full name of the secret with the given prefix and id.
func (s *Store) fullName(prefix, id string) string {
	return prefix + "-" + id
}
