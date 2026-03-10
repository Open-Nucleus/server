package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/config"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	tlsConfig  *tls.Config
}

func New(cfg *config.Config, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
		logger: logger,
	}
}

// WithTLS attaches a TLS configuration to the server.
// If tc is nil, the server runs plain HTTP.
func (s *Server) WithTLS(tc *tls.Config) *Server {
	s.tlsConfig = tc
	if tc != nil {
		s.httpServer.TLSConfig = tc
	}
	return s
}

// Run starts the server and blocks until a shutdown signal is received.
func (s *Server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if s.tlsConfig != nil {
			s.logger.Info("starting HTTPS server", "addr", s.httpServer.Addr)
			// TLS cert/key already loaded into TLSConfig, pass empty strings.
			if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		} else {
			s.logger.Info("starting HTTP server", "addr", s.httpServer.Addr)
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			}
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	}
}
