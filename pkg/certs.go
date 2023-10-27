package courier

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trisacrypto/courier/pkg/api/v1"
)

// StoreCertificatePassword stores the password for an encrypted certificate and
// returns a 204 No Content response.
func (s *Server) StoreCertificatePassword(c *gin.Context) {
	var (
		err error
		req *api.StorePasswordRequest
	)

	// Parse the request body
	req = &api.StorePasswordRequest{}
	if err := c.BindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse(err))
		return
	}

	// Password is required
	if req.Password == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse("missing password in request"))
		return
	}

	// Store the password
	if err = s.store.UpdatePassword(c.Param("id"), []byte(req.Password)); err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
		return
	}

	// Return 204 No Content
	c.Status(http.StatusNoContent)
}
