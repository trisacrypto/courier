package store

// Store is a generic interface for storing and retrieving data.
type Store interface {
	Close() error
	PasswordStore
	CertificateStore
}

// PasswordStore is a generic interface for storing and retrieving passwords.
type PasswordStore interface {
	GetPassword(name string) ([]byte, error)
	UpdatePassword(name string, password []byte) error
}

// CertificateStore is a generic interface for storing and retrieving certificates.
type CertificateStore interface {
	GetCertificate(name string) ([]byte, error)
	UpdateCertificate(name string, cert []byte) error
}
