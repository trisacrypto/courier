package local

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/trisacrypto/courier/pkg/config"
	"github.com/trisacrypto/courier/pkg/store"
)

const (
	passwordFile    = "pkcs12.password"
	certificateFile = "certificate"
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
	return s.load(s.fullPath(store.PasswordPrefix, id))
}

// UpdatePassword updates a password by id in the local storage backend. If the
// password does not exist, it is created. Otherwise, it is overwritten.
func (s *Store) UpdatePassword(ctx context.Context, id string, password []byte) (err error) {
	s.Lock()
	defer s.Unlock()
	return s.store(s.fullPath(store.PasswordPrefix, id), passwordFile, password)
}

//===========================================================================
// Certificate Methods
//===========================================================================

// GetCertificate retrieves a certificate by id from the local storage backend.
func (s *Store) GetCertificate(ctx context.Context, name string) (cert []byte, err error) {
	s.RLock()
	defer s.RUnlock()
	return s.load(s.fullPath(store.CertificatePrefix, name))
}

// UpdateCertificate updates a certificate in the local storage backend.
func (s *Store) UpdateCertificate(ctx context.Context, name string, cert []byte) (err error) {
	s.Lock()
	defer s.Unlock()
	return s.store(s.fullPath(store.CertificatePrefix, name), certificateFile, cert)
}

//===========================================================================
// Helper methods
//===========================================================================

// fullPath returns the full path to an archive file in the local storage backend.
func (s *Store) fullPath(prefix, name string) string {
	return filepath.Join(s.path, prefix+"-"+name+".zip")
}

// load returns file data by archive path from the local storage
func (s *Store) load(path string) (data []byte, err error) {
	var archive *zip.ReadCloser
	if archive, err = zip.OpenReader(path); err != nil {
		if os.IsNotExist(err) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	defer archive.Close()

	// Load the file from the archive
	if len(archive.File) == 0 {
		return nil, store.ErrNotFound
	}

	var reader io.ReadCloser
	if reader, err = archive.File[0].Open(); err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// store saves file data to an archive and file name in the local storage
func (s *Store) store(path, name string, data []byte) (err error) {
	var archive *os.File
	if archive, err = os.Create(path); err != nil {
		return err
	}
	defer archive.Close()

	// Write the file to the archive
	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	var writer io.Writer
	if writer, err = zipWriter.Create(name); err != nil {
		return err
	}

	if _, err = writer.Write(data); err != nil {
		return err
	}

	return nil
}
