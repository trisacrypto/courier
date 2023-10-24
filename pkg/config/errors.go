package config

import "errors"

var (
	ErrMissingBindAddr   = errors.New("invalid configuration: missing bindaddr")
	ErrMissingServerMode = errors.New("invalid configuration: missing server mode (debug, release, test)")
	ErrMissingCertPaths  = errors.New("invalid configuration: missing cert path or pool path")
	ErrTLSNotConfigured  = errors.New("cannot create TLS configuration in insecure mode")
)
