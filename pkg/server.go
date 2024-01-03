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
	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/courier/pkg/api/v1"
	"github.com/trisacrypto/courier/pkg/config"
	"github.com/trisacrypto/courier/pkg/store"
	"github.com/trisacrypto/courier/pkg/store/gcloud"
	"github.com/trisacrypto/courier/pkg/store/local"
)

// New creates a new server object from configuration but does not serve it yet.
func New(conf config.Config) (s *Server, err error) {
	// Load config from environment if it's empty
	if conf.IsZero() {
		if conf, err = config.New(); err != nil {
			return nil, err
		}
	}

	// Create the server object
	s = &Server{
		conf:  conf,
		echan: make(chan error, 1),
	}

	// Open the store
	switch {
	case conf.LocalStorage.Enabled:
		if s.store, err = local.Open(conf.LocalStorage); err != nil {
			return nil, err
		}
	case conf.GCPSecretManager.Enabled:
		if s.store, err = gcloud.Open(conf.GCPSecretManager); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("no storage backend configured")
	}

	// Create the router
	gin.SetMode(conf.Mode)
	s.router = gin.New()
	if err = s.setupRoutes(); err != nil {
		return nil, err
	}

	// Create the http server
	s.srv = &http.Server{
		Addr:    conf.BindAddr,
		Handler: s.router,
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
	conf    config.Config
	srv     *http.Server
	router  *gin.Engine
	store   store.Store
	started time.Time
	healthy bool
	url     string
	echan   chan error
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

	if err = s.srv.Shutdown(ctx); err != nil {
		return err
	}

	// TODO: Close the stores

	log.Debug().Msg("successfully shut down courier server")
	return nil
}

// Setup the routes for the courier service.
func (s *Server) setupRoutes() (err error) {
	middlewares := []gin.HandlerFunc{
		gin.Logger(),
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

// Set the healthy status of the server.
func (s *Server) SetHealthy(healthy bool) {
	s.Lock()
	s.healthy = healthy
	s.Unlock()
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
