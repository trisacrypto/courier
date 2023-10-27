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

	// Decode the certificate data
	var data []byte
	if data, err = base64.StdEncoding.DecodeString(req.Base64Certificate); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse(err))
		return
	}

	// If encryption is enabled, retrieve the pkcs12 password from the store
	var password string
	if !req.NoDecrypt {
		var password []byte
		if password, err = s.store.GetPassword(id); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, api.ErrorResponse("pkcs12 password not found, unable to decrypt certificate"))
				return
			}

			c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
			return
		}
	}

	// Make sure we can decode the certificate, decrypting if necessary
	var sz *trust.Serializer
	if sz, err = trust.NewSerializer(req.NoDecrypt, password); err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
		return
	}

	var provider *trust.Provider
	if provider, err = sz.Extract(data); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse(err))
		return
	}

	// Store the certificate
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
	if err = s.store.UpdatePassword(c.Param("id"), []byte(req.Password)); err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
		return
	}

	// Return 204 No Content
	c.Status(http.StatusNoContent)
}
