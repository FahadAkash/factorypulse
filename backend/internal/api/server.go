package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/FahadAkash/factorypulse/backend/internal/incident"
	"github.com/FahadAkash/factorypulse/backend/internal/ws"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	apiHandler *APIHandler
}

func NewServer(port string, hub *ws.Hub, incidents *incident.Service, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	s := &Server{
		logger:     logger,
		apiHandler: NewAPIHandler(hub, incidents),
	}

	// Register core probes
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)

	// Register REST & WS routes
	s.apiHandler.RegisterRoutes(mux)

	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpServer.Handler.ServeHTTP(w, r)
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("HTTP server starting", "addr", s.httpServer.Addr)

	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("HTTP server shutting down gracefully")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-errChan:
		return fmt.Errorf("http server failure: %w", err)
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "UP",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "READY",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
