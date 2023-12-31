package courier

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/courier/pkg/api/v1"
	"github.com/trisacrypto/courier/pkg/config"
	"github.com/trisacrypto/courier/pkg/logger"
	"github.com/trisacrypto/courier/pkg/o11y"
	"github.com/trisacrypto/courier/pkg/store"
	"github.com/trisacrypto/courier/pkg/store/gcloud"
	"github.com/trisacrypto/courier/pkg/store/local"
)

func init() {
	// Initializes zerolog with our default logging requirements
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = logger.GCPFieldKeyTime
	zerolog.MessageFieldName = logger.GCPFieldKeyMsg
	zerolog.DurationFieldInteger = false
	zerolog.DurationFieldUnit = time.Millisecond

	// Add the severity hook for GCP logging
	var gcpHook logger.SeverityHook
	log.Logger = zerolog.New(os.Stdout).Hook(gcpHook).With().Timestamp().Logger()
}

// New creates a new server object from configuration but does not serve it yet.
func New(conf config.Config) (s *Server, err error) {
	// Load config from environment if it's empty
	if conf.IsZero() {
		if conf, err = config.New(); err != nil {
			return nil, err
		}
	}

	// Setup our logging config first thing
	zerolog.SetGlobalLevel(conf.GetLogLevel())
	if conf.ConsoleLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Create the server object
	s = &Server{
		conf:  conf,
		echan: make(chan error, 1),
	}

	// Open the store
	if !s.conf.Maintenance {
		switch {
		case s.conf.LocalStorage.Enabled:
			if s.store, err = local.Open(s.conf.LocalStorage); err != nil {
				return nil, err
			}
		case s.conf.GCPSecretManager.Enabled:
			if s.store, err = gcloud.Open(s.conf.GCPSecretManager); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("no storage backend configured")
		}
	}

	// Create the router
	gin.SetMode(conf.Mode)
	s.router = gin.New()
	s.router.RedirectTrailingSlash = true
	s.router.RedirectFixedPath = false
	s.router.HandleMethodNotAllowed = true
	s.router.ForwardedByClientIP = true
	s.router.UseRawPath = false
	s.router.UnescapePathValues = true

	if err = s.setupRoutes(); err != nil {
		return nil, err
	}

	// Create the http server
	s.srv = &http.Server{
		Addr:              conf.BindAddr,
		Handler:           s.router,
		ErrorLog:          nil,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	// Use TLS if configured
	if !conf.MTLS.Insecure {
		if s.srv.TLSConfig, err = conf.MTLS.ParseTLSConfig(); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Server defines the courier service and its webhook handlers.
type Server struct {
	sync.RWMutex
	conf    config.Config // Primary source of truth for server configuration
	srv     *http.Server  // The HTTP server for handling requests
	router  *gin.Engine   // The gin router for muxing requests to handlers
	store   store.Store   // Manages certificate and password storage
	healthy bool          // Indicates that the service is online and healthy
	ready   bool          // Indicates that the service is ready to accept requests
	started time.Time     // The timestamp the server was started (for uptime)
	url     string        // The endpoint that the server is hosted on
	echan   chan error    // Sending errors on this channel stops the server
}

// Serve API requests.
func (s *Server) Serve() (err error) {
	// Catch OS signals for graceful shutdowns
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		s.echan <- s.Shutdown()
	}()

	// Set healthy status
	s.SetHealthy(true)

	// Create the listen socket
	var sock net.Listener
	if sock, err = net.Listen("tcp", s.conf.BindAddr); err != nil {
		return err
	}

	// Set the URL from the socket
	s.SetURL(sock)
	s.started = time.Now()

	// Serve the API
	go func() {
		if err = s.srv.Serve(sock); err != nil && err != http.ErrServerClosed {
			s.echan <- err
		}
	}()

	s.SetReady(true)
	log.Info().Str("listen", s.url).Str("version", Version()).Msg("courier server started")

	// Wait for shutdown or an error
	if err = <-s.echan; err != nil {
		return err
	}
	return nil
}

// Shutdown the server gracefully.
func (s *Server) Shutdown() (err error) {
	log.Info().Msg("gracefully shutting down courier server")

	s.SetHealthy(false)
	s.srv.SetKeepAlivesEnabled(false)

	// Ensure shutdown happens within 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if serr := s.srv.Shutdown(ctx); serr != nil {
		err = errors.Join(err, serr)
	}

	if !s.conf.Maintenance {
		if serr := s.store.Close(); serr != nil {
			err = errors.Join(err, serr)
		}
	}

	log.Debug().Err(err).Msg("shut down courier server")
	return err
}

// Setup the routes for the courier service.
func (s *Server) setupRoutes() (err error) {
	// Kubernetes probe endpoints -- add routes before middleware to ensure these
	// endpoints are not logged or subject to other handling that may harm correctness
	s.router.GET("/healthz", s.Healthz)
	s.router.GET("/livez", s.Healthz)
	s.router.GET("/readyz", s.Readyz)

	// Add prometheus metrics collector endpoint before middleware is added
	s.router.GET("/metrics", o11y.Prometheus())

	middlewares := []gin.HandlerFunc{
		logger.GinLogger("courier", Version()),
		o11y.Metrics(),
		gin.Recovery(),
		s.Available(),
	}

	// Add the middlewares to the router
	s.router.Use(middlewares...)

	// API routes
	v1 := s.router.Group("/v1")
	{
		// Status route
		v1.GET("/status", s.Status)

		// Certificate routes
		certs := v1.Group("/certs")
		{
			certs.POST("/:id", s.StoreCertificate)
			certs.POST("/:id/pkcs12password", s.StoreCertificatePassword)
		}
	}

	// Not found and method not allowed routes
	s.router.NoRoute(api.NotFound)
	s.router.NoMethod(api.MethodNotAllowed)
	return nil
}

// Set the URL of the server from the socket
func (s *Server) SetURL(sock net.Listener) {
	s.Lock()
	sockAddr := sock.Addr().String()
	if s.conf.MTLS.Insecure {
		s.url = "http://" + sockAddr
	} else {
		s.url = "https://" + sockAddr
	}
	s.Unlock()
}

//===========================================================================
// Helpers for testing
//===========================================================================

// URL returns the URL of the server.
func (s *Server) URL() string {
	s.RLock()
	defer s.RUnlock()
	return s.url
}

// SetStore directly sets the store for the server.
func (s *Server) SetStore(store store.Store) {
	s.store = store
}
