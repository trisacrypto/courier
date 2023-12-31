package local_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trisacrypto/courier/pkg/config"
	"github.com/trisacrypto/courier/pkg/store"
	"github.com/trisacrypto/courier/pkg/store/local"
)

type localStoreTestSuite struct {
	suite.Suite
	conf  config.LocalStorageConfig
	store *local.Store
}

func (s *localStoreTestSuite) SetupSuite() {
	// Open the storage backend in a temporary directory
	var err error
	path := s.T().TempDir()
	s.conf = config.LocalStorageConfig{
		Enabled: true,
		Path:    path,
	}
	s.store, err = local.Open(s.conf)
	s.NoError(err, "could not open local storage backend")
}

func (s *localStoreTestSuite) TearDownSuite() {
	// Remove the temporary directory
	s.NoError(s.store.Close(), "could not close local storage backend")
	s.NoError(os.RemoveAll(s.conf.Path), "could not remove temporary directory")
}

func TestLocalStore(t *testing.T) {
	suite.Run(t, new(localStoreTestSuite))
}

func (s *localStoreTestSuite) TestPasswordStore() {
	require := s.Require()
	ctx := context.Background()

	// Try to get a password that does not exist
	_, err := s.store.GetPassword(ctx, "does-not-exist")
	require.ErrorIs(err, store.ErrNotFound, "should return error if password does not exist")

	// Create a password
	password := []byte("password")
	err = s.store.UpdatePassword(ctx, "password_id", password)
	require.NoError(err, "should be able to create a password")

	// Get the password
	actual, err := s.store.GetPassword(ctx, "password_id")
	require.NoError(err, "should be able to get a password")
	require.Equal(password, actual, "wrong password returned")
}

func (s *localStoreTestSuite) TestCertificateStore() {
	require := s.Require()
	ctx := context.Background()

	// Try to get a certificate that does not exist
	_, err := s.store.GetCertificate(ctx, "does-not-exist")
	require.ErrorIs(err, store.ErrNotFound, "should return error if certificate does not exist")

	// Create a certificate
	cert := []byte("certificate")
	err = s.store.UpdateCertificate(ctx, "certificate_id", cert)
	require.NoError(err, "should be able to create a certificate")

	// Get the certificate
	actual, err := s.store.GetCertificate(ctx, "certificate_id")
	require.NoError(err, "should be able to get a certificate")
	require.Equal(cert, actual, "wrong certificate returned")
}
