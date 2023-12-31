package store

import (
	"context"
	"io"
)

const (
	PasswordPrefix    = "pkcs12"
	CertificatePrefix = "certificate"
)

// Store is a generic interface for storing and retrieving data.
type Store interface {
	io.Closer
	PasswordStore
	CertificateStore
}

// PasswordStore is a generic interface for storing and retrieving passwords.
type PasswordStore interface {
	GetPassword(ctx context.Context, name string) ([]byte, error)
	UpdatePassword(ctx context.Context, name string, password []byte) error
}

// CertificateStore is a generic interface for storing and retrieving certificates.
type CertificateStore interface {
	GetCertificate(ctx context.Context, name string) ([]byte, error)
	UpdateCertificate(ctx context.Context, name string, cert []byte) error
}
