package api_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/courier/pkg/api/v1"
)

func TestJoinStatusErrors(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		err := api.JoinStatusErrors(0, 0, nil)
		require.NoError(t, err, "expected a nil error returned")

		err = api.JoinStatusErrors(0, 0, nil, nil, nil, nil, nil, nil)
		require.NoError(t, err, "expected a nil error returned for multiple nil errors")
	})

	t.Run("SingleStatusError", func(t *testing.T) {
		err := api.JoinStatusErrors(1, 421*time.Millisecond, api.NewStatusError(http.StatusServiceUnavailable, ""))
		require.Error(t, err, "expected error to be returned")

		_, ok := err.(*api.StatusError)
		require.True(t, ok, "expected error to be a status error, not a multi status error")
		require.EqualError(t, err, "[503]: Service Unavailable")
	})

	t.Run("SingleError", func(t *testing.T) {
		err := api.JoinStatusErrors(1, 421*time.Millisecond, errors.New("something went wrong"))
		require.Error(t, err, "expected error to be returned")

		_, ok := err.(*api.StatusError)
		require.False(t, ok, "expected error to not be a status error")
		require.EqualError(t, err, "something went wrong")
	})

	t.Run("MultiStatusErrors", func(t *testing.T) {
		err := api.JoinStatusErrors(3, 1829*time.Millisecond,
			api.NewStatusError(http.StatusUnauthorized, ""),
			api.NewStatusError(http.StatusServiceUnavailable, ""),
			api.NewStatusError(http.StatusInsufficientStorage, ""),
		)
		require.Error(t, err, "expected error to be returned")

		_, ok := err.(*api.MultiStatusError)
		require.True(t, ok, "expected error to be a multi-status error")
		require.EqualError(t, err, "after 3 attempts: [507]: Insufficient Storage")
	})

	t.Run("MultiErrors", func(t *testing.T) {
		err := api.JoinStatusErrors(2, 727*time.Millisecond,
			errors.New("oopsie"), errors.New("something went wrong"),
		)
		require.Error(t, err, "expected error to be returned")

		_, ok := err.(*api.MultiStatusError)
		require.True(t, ok, "expected error to be a multi-status error")
		require.EqualError(t, err, "after 2 attempts: something went wrong")
	})

	t.Run("Mixed", func(t *testing.T) {
		err := api.JoinStatusErrors(2, 3217*time.Millisecond,
			api.NewStatusError(http.StatusServiceUnavailable, ""),
			errors.New("something went wrong"),
		)
		require.Error(t, err, "expected error to be returned")

		_, ok := err.(*api.MultiStatusError)
		require.True(t, ok, "expected error to be a multi-status error")
		require.EqualError(t, err, "after 2 attempts: something went wrong")
	})

	t.Run("Deduplication", func(t *testing.T) {
		err := api.JoinStatusErrors(3, 2451*time.Millisecond,
			api.NewStatusError(http.StatusServiceUnavailable, ""),
			api.NewStatusError(http.StatusServiceUnavailable, ""),
			api.NewStatusError(http.StatusServiceUnavailable, ""),
		)
		require.Error(t, err, "expected error to be returned")

		_, ok := err.(*api.StatusError)
		require.True(t, ok, "expected error to be a status error")
		require.EqualError(t, err, "[503]: Service Unavailable")
	})

	t.Run("MultiDeduplication", func(t *testing.T) {
		err := api.JoinStatusErrors(5, 3257*time.Millisecond,
			api.NewStatusError(http.StatusUnauthorized, ""),
			api.NewStatusError(http.StatusServiceUnavailable, ""),
			api.NewStatusError(http.StatusUnauthorized, ""),
			api.NewStatusError(http.StatusInsufficientStorage, ""),
			api.NewStatusError(http.StatusServiceUnavailable, ""),
		)
		require.Error(t, err, "expected error to be returned")

		_, ok := err.(*api.MultiStatusError)
		require.True(t, ok, "expected error to be a multi-status error")
		require.EqualError(t, err, "after 5 attempts: [507]: Insufficient Storage")
	})
}

func TestMultiStatusError(t *testing.T) {
	testCases := []struct {
		err      *api.MultiStatusError
		expected string
	}{
		{
			&api.MultiStatusError{
				Attempts: 1,
				Delay:    585 * time.Millisecond,
				Errs: []error{
					&api.StatusError{
						Code: http.StatusInternalServerError,
						Err:  http.StatusText(http.StatusInternalServerError),
					},
				},
			},
			"after 1 attempts: [500]: Internal Server Error",
		},
	}

	for i, tc := range testCases {
		require.EqualError(t, tc.err, tc.expected, "test case %d failed", i)
	}
}
