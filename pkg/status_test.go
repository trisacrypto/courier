package courier_test

import "context"

func (s *courierTestSuite) TestStatus() {
	require := s.Require()

	// Make a request to the status endpoint
	status, err := s.client.Status(context.Background())
	require.NoError(err, "could not get status from server")

	// Check that the status is as expected
	require.Equal("ok", status.Status, "status should be ok")
	require.NotEmpty(status.Uptime, "uptime missing from response")
	require.NotEmpty(status.Version, "version missing from response")
}
