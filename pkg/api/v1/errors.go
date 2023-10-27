package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	unsuccessful        = Reply{Success: false}
	notFound            = Reply{Success: false, Error: "resource not found"}
	notAllowed          = Reply{Success: false, Error: "method not allowed"}
	ErrEndpointRequired = errors.New("endpoint is required")
	ErrIDRequired       = errors.New("missing ID in request")
)

func NewStatusError(code int, err string) error {
	return &StatusError{Code: code, Err: err}
}

type StatusError struct {
	Code int
	Err  string
}

func (e StatusError) Error() string {
	return fmt.Sprintf("[%d]: %s", e.Code, e.Err)
}

// ErrorResponse constructs an new response from the error or returns a success: false.
func ErrorResponse(err interface{}) Reply {
	if err == nil {
		return unsuccessful
	}

	rep := Reply{Success: false}
	switch err := err.(type) {
	case error:
		rep.Error = err.Error()
	case string:
		rep.Error = err
	case fmt.Stringer:
		rep.Error = err.String()
	case json.Marshaler:
		data, e := err.MarshalJSON()
		if e != nil {
			panic(err)
		}
		rep.Error = string(data)
	default:
		rep.Error = "unhandled error response"
	}

	return rep
}

// NotFound returns a standard 404 response.
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, notFound)
}

// MethodNotAllowed returns a standard 405 response.
func MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, notAllowed)
}
