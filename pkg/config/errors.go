package config

import "errors"

var (
	ErrMissingBindAddr           = errors.New("invalid configuration: missing bindaddr")
	ErrMissingServerMode         = errors.New("invalid configuration: missing server mode (debug, release, test)")
	ErrMissingCertPaths          = errors.New("invalid configuration: missing cert path or pool path")
	ErrTLSNotConfigured          = errors.New("cannot create TLS configuration in insecure mode")
	ErrMissingLocalPath          = errors.New("invalid configuration: missing path for local storage")
	ErrNoStorageEnabled          = errors.New("invalid configuration: must enable either local storage or secret manager storage")
	ErrMultipleStorageEnabled    = errors.New("invalid configuration: cannot enable both local storage and secret manager storage")
	ErrMissingSecretsCredentials = errors.New("invalid configuration: missing credentials for secret manager storage")
	ErrMissingSecretsProject     = errors.New("invalid configuration: missing project name for secret manager storage")
)
