package secrets

import "errors"

var (
	ErrSecretNotFound    = errors.New("secret not found")
	ErrPayloadTooLarge   = errors.New("secret payload too large")
	ErrPermissionsDenied = errors.New("secret access denied")
)
