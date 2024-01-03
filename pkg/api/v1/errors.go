package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	unsuccessful        = Reply{Success: false}
	notFound            = Reply{Success: false, Error: "resource not found"}
	notAllowed          = Reply{Success: false, Error: "method not allowed"}
	ErrEndpointRequired = errors.New("endpoint is required")
	ErrIDRequired       = errors.New("missing ID in request")
	ErrInvalidRetries   = errors.New("number of retries must be zero or more")
)

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

func NewStatusError(code int, err string) error {
	if err == "" {
		err = http.StatusText(code)
	}
	return &StatusError{Code: code, Err: err}
}

type StatusError struct {
	Code int
	Err  string
}

func (e StatusError) Error() string {
	return fmt.Sprintf("[%d]: %s", e.Code, e.Err)
}

// Deduplicates status errors and creates a multi-status error to return. Removes nil
// errors and returns nil if all errs are nil. If only one errors is returned, return
// that error instead of a multierror (e.g. if all responses have the same status code).
func JoinStatusErrors(attempts int, delay time.Duration, errs ...error) error {
	err := &MultiStatusError{
		Errs:     make([]error, 0),
		Attempts: attempts,
	}

	seen := make(map[string]struct{})
	for _, e := range errs {
		if e == nil {
			continue
		}

		if _, ok := seen[e.Error()]; ok {
			continue
		}

		err.Errs = append(err.Errs, e)
		seen[e.Error()] = struct{}{}
	}

	switch len(err.Errs) {
	case 0:
		return nil
	case 1:
		return err.Errs[0]
	default:
		return err
	}
}

type MultiStatusError struct {
	Errs     []error
	Attempts int
	Delay    time.Duration
}

func (e *MultiStatusError) Error() string {
	return fmt.Sprintf("after %d attempts: %s", e.Attempts, e.Last())
}

func (e *MultiStatusError) Last() error {
	if len(e.Errs) > 0 {
		return e.Errs[len(e.Errs)-1]
	}
	return nil
}

func (e *MultiStatusError) Unwrap() []error {
	return e.Errs
}

// NotFound returns a standard 404 response.
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, notFound)
}

// MethodNotAllowed returns a standard 405 response.
func MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, notAllowed)
}
