package httpserver

import (
	"context"
	"net/http"
	"time"
	"os"
	"os/signal"
	"syscall"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server
type Server struct {
	srv *http.Server
}

// New creates a new server with the given address and mux
func New(addr string, mux *http.ServeMux) *Server {
	return &Server{
		srv: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Run starts the server and blocks until shutdown
func (s *Server) Run(ctx context.Context) error {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server_start_failed", "err", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.srv.Shutdown(shutdownCtx)
}

// DefaultMux creates a mux with health endpoints
func DefaultMux(ready func() bool) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if ready() { w.WriteHeader(200); return }
		w.WriteHeader(503)
	})
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

// RunMain starts the server with health endpoints and graceful shutdown
func RunMain(addr string, ready func() bool) {
	mux := DefaultMux(ready)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := New(addr, mux).Run(ctx); err != nil {
		slog.Error("graceful_shutdown_failed", "err", err)
	}
}

// IntegrateWithExistingServer adds health endpoints to an existing mux
func IntegrateWithExistingServer(existingMux *http.ServeMux, ready func() bool) {
	existingMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	existingMux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if ready() { w.WriteHeader(200); return }
		w.WriteHeader(503)
	})
	existingMux.Handle("/metrics", promhttp.Handler())
}
