package config

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/rotationalio/confire"
	"github.com/rs/zerolog"
	"github.com/trisacrypto/courier/pkg/logger"
	"github.com/trisacrypto/trisa/pkg/trust"
)

type Config struct {
	Maintenance      bool                `default:"false"`
	BindAddr         string              `split_words:"true" default:":8842"`
	Mode             string              `split_words:"true" default:"release"`
	LogLevel         logger.LevelDecoder `split_words:"true" default:"info"`
	ConsoleLog       bool                `split_words:"true" default:"false"`
	MTLS             MTLSConfig          `split_words:"true"`
	LocalStorage     LocalStorageConfig  `split_words:"true"`
	GCPSecretManager GCPSecretsConfig    `split_words:"true"`
	processed        bool
}

type MTLSConfig struct {
	Insecure bool   `split_words:"true" default:"true"`
	CertPath string `split_words:"true"`
	PoolPath string `split_words:"true"`
	pool     *x509.CertPool
	cert     tls.Certificate
}

type LocalStorageConfig struct {
	Enabled bool   `split_words:"true" default:"false"`
	Path    string `split_words:"true"`
}

type GCPSecretsConfig struct {
	Enabled     bool   `split_words:"true" default:"false"`
	Credentials string `split_words:"true"`
	Project     string `split_words:"true"`
}

// Create a new Config struct using values from the environment prefixed with COURIER.
func New() (conf Config, err error) {
	if err = confire.Process("courier", &conf); err != nil {
		return conf, err
	}

	conf.processed = true
	return conf, nil
}

// Return true if the configuration has not been processed (e.g. not loaded from the
// environment or configuration file).
func (c Config) IsZero() bool {
	return !c.processed
}

// Mark a configuration as processed, for cases where the configuration is manually
// created (e.g. in tests).
func (c Config) Mark() (Config, error) {
	if err := c.Validate(); err != nil {
		return c, err
	}
	c.processed = true
	return c, nil
}

// Validate the configuration.
func (c Config) Validate() (err error) {
	if c.BindAddr == "" {
		return ErrMissingBindAddr
	}

	if c.Mode == "" {
		return ErrMissingServerMode
	}

	if err = c.MTLS.Validate(); err != nil {
		return err
	}

	if !c.LocalStorage.Enabled && !c.GCPSecretManager.Enabled {
		return ErrNoStorageEnabled
	}

	if c.LocalStorage.Enabled && c.GCPSecretManager.Enabled {
		return ErrMultipleStorageEnabled
	}

	if err = c.LocalStorage.Validate(); err != nil {
		return err
	}

	if err = c.GCPSecretManager.Validate(); err != nil {
		return err
	}

	return nil
}

// Parse and return the zerolog log level for configuring global logging.
func (c Config) GetLogLevel() zerolog.Level {
	return zerolog.Level(c.LogLevel)
}

func (c *MTLSConfig) Validate() error {
	if c.Insecure {
		return nil
	}

	if c.CertPath == "" || c.PoolPath == "" {
		return ErrMissingCertPaths
	}

	return nil
}

func (c *MTLSConfig) ParseTLSConfig() (_ *tls.Config, err error) {
	if c.Insecure {
		return nil, ErrTLSNotConfigured
	}

	var certPool *x509.CertPool
	if certPool, err = c.GetCertPool(); err != nil {
		return nil, err
	}

	var cert tls.Certificate
	if cert, err = c.GetCert(); err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP521,
			tls.CurveP384,
			tls.CurveP256,
		},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  certPool,
	}, nil
}

func (c *MTLSConfig) GetCertPool() (_ *x509.CertPool, err error) {
	if c.pool == nil {
		if err = c.load(); err != nil {
			return nil, err
		}
	}
	return c.pool, nil
}

func (c *MTLSConfig) GetCert() (_ tls.Certificate, err error) {
	if len(c.cert.Certificate) == 0 {
		if err = c.load(); err != nil {
			return c.cert, err
		}
	}
	return c.cert, nil
}

func (c *MTLSConfig) load() (err error) {
	var sz *trust.Serializer
	if sz, err = trust.NewSerializer(false); err != nil {
		return err
	}

	var pool trust.ProviderPool
	if pool, err = sz.ReadPoolFile(c.PoolPath); err != nil {
		return err
	}

	var provider *trust.Provider
	if provider, err = sz.ReadFile(c.CertPath); err != nil {
		return err
	}

	if c.pool, err = pool.GetCertPool(false); err != nil {
		return err
	}

	if c.cert, err = provider.GetKeyPair(); err != nil {
		return err
	}

	return nil
}

func (c LocalStorageConfig) Validate() (err error) {
	if !c.Enabled {
		return nil
	}

	if c.Path == "" {
		return ErrMissingLocalPath
	}

	return nil
}

func (c GCPSecretsConfig) Validate() (err error) {
	if !c.Enabled {
		return nil
	}

	if c.Credentials == "" {
		return ErrMissingSecretsCredentials
	}

	if c.Project == "" {
		return ErrMissingSecretsProject
	}

	return nil
}
