package courier

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trisacrypto/courier/pkg/api/v1"
	"github.com/trisacrypto/courier/pkg/store"
	"github.com/trisacrypto/trisa/pkg/trust"
)

// StoreCertificate decodes a base64-encoded certificate in the request, decrypts it
// using the password in the store, and stores the decrypted certificate in the store.
// The NoDecrypt option can be used to skip the decryption and store the certificate in
// its encrypted form.
func (s *Server) StoreCertificate(c *gin.Context) {
	var (
		err error
		req *api.StoreCertificateRequest
	)

	id := c.Param("id")
	ctx := c.Request.Context()

	// Parse the request body
	req = &api.StoreCertificateRequest{}
	if err := c.BindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse(err))
		return
	}

	// Certificate is required
	if req.Base64Certificate == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse("missing certificate in request"))
		return
	}

	// Decode the certificate data from the request
	var data []byte
	if data, err = base64.StdEncoding.DecodeString(req.Base64Certificate); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse(err))
		return
	}

	if !req.NoDecrypt {
		// If decryption is enabled, retrieve the pkcs12 password from the store
		var password []byte
		if password, err = s.store.GetPassword(ctx, id); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, api.ErrorResponse("pkcs12 password not found, unable to decrypt certificate"))
				return
			}

			c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
			return
		}

		// Decrypt the certificate using the password
		var provider *trust.Provider
		if provider, err = trust.Decrypt(data, string(password)); err != nil {
			c.JSON(http.StatusConflict, api.ErrorResponse("failed to decrypt certificate with stored pkcs12 password"))
			return
		}

		// Encode the decrypted certificate for storage
		if data, err = provider.Encode(); err != nil {
			c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
			return
		}
	}

	// Store the certificate data
	if err = s.store.UpdateCertificate(ctx, id, data); err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
		return
	}

	// Return 204 No Content
	c.Status(http.StatusNoContent)
}

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
	if err = s.store.UpdatePassword(c.Request.Context(), c.Param("id"), []byte(req.Password)); err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
		return
	}

	// Return 204 No Content
	c.Status(http.StatusNoContent)
}
