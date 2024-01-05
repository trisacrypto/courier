package courier

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trisacrypto/courier/pkg/api/v1"
)

const (
	serverStatusOK          = "ok"
	serverStatusStopping    = "stopping"
	serverStatusMaintenance = "maintenance"
)

// Status returns the status of the server.
func (s *Server) Status(c *gin.Context) {
	// At this point the status is always OK, the available middleware will handle the
	// stopping status.
	out := &api.StatusReply{
		Status:  serverStatusOK,
		Version: Version(),
		Uptime:  time.Since(s.started).String(),
	}

	c.JSON(http.StatusOK, out)
}

// Available is middleware that uses the healthy boolean to return a service unavailable
// http status code if the server is shutting down. This middleware must be first in the
// chain to ensure that complex handling to slow the shutdown of the server.
func (s *Server) Available() gin.HandlerFunc {
	// The server starts in maintenance mode and doesn't change during runtime, so
	// determine what the unhealthy status string is going to be prior to the closure.
	status := serverStatusStopping
	if s.conf.Maintenance {
		status = serverStatusMaintenance
	}

	return func(c *gin.Context) {
		// Check health status
		if s.conf.Maintenance || !s.IsReady() {
			c.JSON(http.StatusServiceUnavailable, api.StatusReply{
				Status:  status,
				Uptime:  time.Since(s.started).String(),
				Version: Version(),
			})

			// Stop processing the request if the server is not ready
			c.Abort()
			return
		}

		// Continue processing the request
		c.Next()
	}
}
