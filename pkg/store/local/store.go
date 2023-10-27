package local

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/trisacrypto/courier/pkg/config"
	"github.com/trisacrypto/courier/pkg/store"
)

const (
	archiveExt = ".gz"
)

// Open the local storage backend.
func Open(conf config.LocalStorageConfig) (store *Store, err error) {
	store = &Store{
		path: conf.Path,
	}

	// Ensure the path exists
	if err = os.MkdirAll(conf.Path, 0755); err != nil {
		return nil, err
	}

	return store, nil
}

// Store implements the store.Store interface for local storage.
type Store struct {
	sync.RWMutex
	path string
}

var _ store.Store = &Store{}

// Close the local storage backend.
func (s *Store) Close() error {
	return nil
}

//===========================================================================
// Password Methods
//===========================================================================

// GetPassword retrieves a password by id from the local storage backend.
func (s *Store) GetPassword(ctx context.Context, id string) (password []byte, err error) {
	s.RLock()
	defer s.RUnlock()
	return s.readFile(s.fullPath(store.PasswordPrefix, id))
}

// UpdatePassword updates a password by id in the local storage backend. If the
// password does not exist, it is created. Otherwise, it is overwritten.
func (s *Store) UpdatePassword(ctx context.Context, id string, password []byte) (err error) {
	s.Lock()
	defer s.Unlock()
	return s.writeFile(s.fullPath(store.PasswordPrefix, id), password)
}

//===========================================================================
// Certificate Methods
//===========================================================================

// GetCertificate retrieves certificate data by id from the local storage backend.
func (s *Store) GetCertificate(ctx context.Context, name string) (cert []byte, err error) {
	s.RLock()
	defer s.RUnlock()

	// Load the certificate archive into bytes
	if cert, err = os.ReadFile(s.fullPath(store.CertificatePrefix, name)); err != nil {
		if os.IsNotExist(err) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return cert, nil
}

// UpdateCertificate updates certificate data in the local storage backend.
func (s *Store) UpdateCertificate(ctx context.Context, name string, cert []byte) (err error) {
	s.Lock()
	defer s.Unlock()
	return os.WriteFile(s.fullPath(store.CertificatePrefix, name), cert, 0644)
}

//===========================================================================
// Helper methods
//===========================================================================

// fullPath returns the full path to an archive file in the local storage backend.
func (s *Store) fullPath(prefix, name string) string {
	return filepath.Join(s.path, prefix+"-"+name+archiveExt)
}

// read returns file data by archive path from the local storage
func (s *Store) readFile(path string) (data []byte, err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return nil, err
	}

	var reader *gzip.Reader
	if reader, err = gzip.NewReader(f); err != nil {
		return nil, err
	}

	return io.ReadAll(reader)
}

// write saves file data to an archive file in the local storage
func (s *Store) writeFile(path string, data []byte) (err error) {
	// Write the data to the archive
	var b bytes.Buffer
	writer := gzip.NewWriter(&b)
	if _, err = writer.Write(data); err != nil {
		return err
	}

	// Write the archive to the file
	if err = writer.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0644)
}
