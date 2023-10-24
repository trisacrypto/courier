package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/courier/pkg/config"
)

// Define a test environment for the config tests.
var testEnv = map[string]string{
	"COURIER_BIND_ADDR":      ":8080",
	"COURIER_MODE":           "debug",
	"COURIER_MTLS_INSECURE":  "false",
	"COURIER_MTLS_CERT_PATH": "/path/to/cert",
	"COURIER_MTLS_POOL_PATH": "/path/to/pool",
}

func TestConfig(t *testing.T) {
	// Set required environment variables
	prevEnv := curEnv()
	t.Cleanup(func() {
		for key, val := range prevEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	})
	setEnv()

	conf, err := config.New()
	require.NoError(t, err, "could not create config from test environment")
	require.False(t, conf.IsZero(), "config should be processed")

	require.Equal(t, testEnv["COURIER_BIND_ADDR"], conf.BindAddr)
	require.Equal(t, testEnv["COURIER_MODE"], conf.Mode)
	require.False(t, conf.MTLS.Insecure)
	require.Equal(t, testEnv["COURIER_MTLS_CERT_PATH"], conf.MTLS.CertPath)
	require.Equal(t, testEnv["COURIER_MTLS_POOL_PATH"], conf.MTLS.PoolPath)
}

func TestValidate(t *testing.T) {
	t.Run("ValidInsecure", func(t *testing.T) {
		conf := config.Config{
			BindAddr: ":8080",
			Mode:     "debug",
			MTLS: config.MTLSConfig{
				Insecure: true,
			},
		}
		require.NoError(t, conf.Validate(), "insecure config should be valid")
	})

	t.Run("ValidSecure", func(t *testing.T) {
		conf := config.Config{
			BindAddr: ":8080",
			Mode:     "debug",
			MTLS: config.MTLSConfig{
				CertPath: "/path/to/cert",
				PoolPath: "/path/to/pool",
			},
		}
		require.NoError(t, conf.Validate(), "secure config should be valid")
	})

	t.Run("MissingBindAddr", func(t *testing.T) {
		conf := config.Config{
			Mode: "debug",
			MTLS: config.MTLSConfig{
				Insecure: true,
			},
		}
		require.ErrorIs(t, conf.Validate(), config.ErrMissingBindAddr, "config should be invalid")
	})

	t.Run("MissingServerMode", func(t *testing.T) {
		conf := config.Config{
			BindAddr: ":8080",
			MTLS: config.MTLSConfig{
				Insecure: true,
			},
		}
		require.ErrorIs(t, conf.Validate(), config.ErrMissingServerMode, "config should be invalid")
	})

	t.Run("MissingCertPaths", func(t *testing.T) {
		conf := config.Config{
			BindAddr: ":8080",
			Mode:     "debug",
			MTLS: config.MTLSConfig{
				Insecure: false,
			},
		}
		require.ErrorIs(t, conf.Validate(), config.ErrMissingCertPaths, "config should be invalid")
	})
}

// Returns the current environment for the specified keys, or if no keys are specified
// then returns the current environment for all keys in testEnv.
func curEnv(keys ...string) map[string]string {
	env := make(map[string]string)
	if len(keys) > 0 {
		for _, envvar := range keys {
			if val, ok := os.LookupEnv(envvar); ok {
				env[envvar] = val
			}
		}
	} else {
		for key := range testEnv {
			env[key] = os.Getenv(key)
		}
	}

	return env
}

// Sets the environment variable from the testEnv, if no keys are specified, then sets
// all environment variables from the test env.
func setEnv(keys ...string) {
	if len(keys) > 0 {
		for _, key := range keys {
			if val, ok := testEnv[key]; ok {
				os.Setenv(key, val)
			}
		}
	} else {
		for key, val := range testEnv {
			os.Setenv(key, val)
		}
	}
}
