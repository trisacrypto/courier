package courier

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trisacrypto/courier/pkg/api/v1"
)

const (
	serverStatusOK       = "ok"
	serverStatusStopping = "stopping"
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
	return func(c *gin.Context) {
		// Check health status
		s.RLock()
		if !s.healthy {
			c.JSON(http.StatusServiceUnavailable, api.StatusReply{
				Status:  serverStatusStopping,
				Uptime:  time.Since(s.started).String(),
				Version: Version(),
			})

			c.Abort()
			s.RUnlock()
			return
		}
		s.RUnlock()
		c.Next()
	}
}
