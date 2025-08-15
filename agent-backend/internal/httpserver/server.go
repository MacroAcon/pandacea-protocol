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

type Server struct {
	srv *http.Server
}

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

func RunMain(addr string, ready func() bool) {
	mux := DefaultMux(ready)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := New(addr, mux).Run(ctx); err != nil {
		slog.Error("graceful_shutdown_failed", "err", err)
	}
}
