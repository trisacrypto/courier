package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	notFound            = Reply{Success: false, Error: "resource not found"}
	notAllowed          = Reply{Success: false, Error: "method not allowed"}
	ErrEndpointRequired = errors.New("endpoint is required")
)

func NewStatusError(code int, err string) error {
	return StatusError{Code: code, Err: err}
}

type StatusError struct {
	Code int
	Err  string
}

func (e StatusError) Error() string {
	return fmt.Sprintf("[%d]: %s", e.Code, e.Err)
}

// NotFound returns a standard 404 response.
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, notFound)
}

// MethodNotAllowed returns a standard 405 response.
func MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, notAllowed)
}
