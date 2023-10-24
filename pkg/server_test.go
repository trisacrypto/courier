package courier_test

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	courier "github.com/trisacrypto/courier/pkg"
	"github.com/trisacrypto/courier/pkg/api/v1"
	"github.com/trisacrypto/courier/pkg/config"
)

// The courier test suite allows us to test the courier API by making actual requests
// to an in-memory server.
type courierTestSuite struct {
	suite.Suite
	courier *courier.Server
	client  api.CourierClient
}

func (s *courierTestSuite) SetupSuite() {
	require := s.Require()

	// Configuration to start a fully functional server for localhost testing.
	conf, err := config.Config{
		BindAddr: "127.0.0.1:0",
		Mode:     gin.TestMode,
		MTLS: config.MTLSConfig{
			Insecure: true,
		},
	}.Mark()
	require.NoError(err, "could not create test configuration")

	// Create the server
	s.courier, err = courier.New(conf)
	require.NoError(err, "could not create test server")

	// Start the server, which will run for the duration of the test suite
	go s.courier.Serve()

	// Wait for the server to start serving the API
	time.Sleep(500 * time.Millisecond)

	// Create an API client to use in tests
	url := s.courier.URL()
	s.client, err = api.New(url)
	require.NoError(err, "could not create test client")
}

func (s *courierTestSuite) TearDownSuite() {
	require := s.Require()
	require.NoError(s.courier.Shutdown(), "could not shutdown test server in suite teardown")
}

func TestCourier(t *testing.T) {
	suite.Run(t, new(courierTestSuite))
}
