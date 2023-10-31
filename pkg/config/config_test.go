package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trisacrypto/courier/pkg/config"
)

// Define a test environment for the config tests.
var testEnv = map[string]string{
	"COURIER_BIND_ADDR":                  ":8080",
	"COURIER_MODE":                       "debug",
	"COURIER_MTLS_INSECURE":              "false",
	"COURIER_MTLS_CERT_PATH":             "/path/to/cert",
	"COURIER_MTLS_POOL_PATH":             "/path/to/pool",
	"COURIER_LOCAL_STORAGE_ENABLED":      "true",
	"COURIER_LOCAL_STORAGE_PATH":         "/path/to/storage",
	"COURIER_SECRET_MANAGER_ENABLED":     "true",
	"COURIER_SECRET_MANAGER_CREDENTIALS": "test-credentials",
	"COURIER_SECRET_MANAGER_PROJECT":     "test-project",
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
	require.True(t, conf.LocalStorage.Enabled)
	require.Equal(t, testEnv["COURIER_LOCAL_STORAGE_PATH"], conf.LocalStorage.Path)
	require.True(t, conf.SecretManager.Enabled)
	require.Equal(t, testEnv["COURIER_SECRET_MANAGER_CREDENTIALS"], conf.SecretManager.Credentials)
	require.Equal(t, testEnv["COURIER_SECRET_MANAGER_PROJECT"], conf.SecretManager.Project)
}

func TestValidate(t *testing.T) {
	t.Run("ValidInsecure", func(t *testing.T) {
		conf := config.Config{
			BindAddr: ":8080",
			Mode:     "debug",
			MTLS: config.MTLSConfig{
				Insecure: true,
			},
			LocalStorage: config.LocalStorageConfig{
				Enabled: true,
				Path:    "/path/to/storage",
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
			LocalStorage: config.LocalStorageConfig{
				Enabled: true,
				Path:    "/path/to/storage",
			},
		}
		require.NoError(t, conf.Validate(), "secure config should be valid")
	})

	t.Run("ValidSecretManager", func(t *testing.T) {
		conf := config.Config{
			BindAddr: ":8080",
			Mode:     "debug",
			MTLS: config.MTLSConfig{
				Insecure: true,
			},
			SecretManager: config.SecretsConfig{
				Enabled:     true,
				Credentials: "test-credentials",
				Project:     "test-project",
			},
		}
		require.NoError(t, conf.Validate(), "secret manager config should be valid")
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

	t.Run("MissingStorage", func(t *testing.T) {
		conf := config.Config{
			BindAddr: ":8080",
			Mode:     "debug",
			MTLS: config.MTLSConfig{
				Insecure: true,
			},
		}
		require.ErrorIs(t, conf.Validate(), config.ErrNoStorageEnabled, "config should be invalid")
	})

	t.Run("MultipleStorage", func(t *testing.T) {
		conf := config.Config{
			BindAddr: ":8080",
			Mode:     "debug",
			MTLS: config.MTLSConfig{
				Insecure: true,
			},
			LocalStorage: config.LocalStorageConfig{
				Enabled: true,
				Path:    "/path/to/storage",
			},
			SecretManager: config.SecretsConfig{
				Enabled:     true,
				Credentials: "test-credentials",
				Project:     "test-project",
			},
		}
		require.ErrorIs(t, conf.Validate(), config.ErrMultipleStorageEnabled, "config should be invalid")
	})

	t.Run("MissingLocalPath", func(t *testing.T) {
		conf := config.Config{
			BindAddr: ":8080",
			Mode:     "debug",
			MTLS: config.MTLSConfig{
				Insecure: true,
			},
			LocalStorage: config.LocalStorageConfig{
				Enabled: true,
			},
		}
		require.ErrorIs(t, conf.Validate(), config.ErrMissingLocalPath, "config should be invalid")
	})
}

func TestValidateSecretConfig(t *testing.T) {
	t.Run("ValidSecretConfig", func(t *testing.T) {
		conf := config.SecretsConfig{
			Enabled:     true,
			Credentials: "test-credentials",
			Project:     "test-project",
		}
		require.NoError(t, conf.Validate(), "secret config should be valid")
	})

	t.Run("ValidDisabled", func(t *testing.T) {
		conf := config.SecretsConfig{}
		require.NoError(t, conf.Validate(), "expected disabled secret config to be valid")
	})

	t.Run("MissingCredentials", func(t *testing.T) {
		conf := config.SecretsConfig{
			Enabled: true,
			Project: "test-project",
		}
		require.ErrorIs(t, conf.Validate(), config.ErrMissingSecretsCredentials, "config should be invalid")
	})

	t.Run("MissingProject", func(t *testing.T) {
		conf := config.SecretsConfig{
			Enabled:     true,
			Credentials: "test-credentials",
		}
		require.ErrorIs(t, conf.Validate(), config.ErrMissingSecretsProject, "config should be invalid")
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
