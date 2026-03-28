package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"
)

// Constants for HTTP trigger server timeouts
const (
	defaultTriggerReadTimeout  = 5 * time.Second
	defaultTriggerWriteTimeout = 10 * time.Second
	defaultTriggerIdleTimeout  = 30 * time.Second
	triggerShutdownTimeout     = 10 * time.Second
)

// Global trigger server state
var (
	triggerServer *http.Server
	triggerCancel context.CancelFunc
)

// TriggerResponse represents the HTTP trigger response
type TriggerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// StartTriggerServer starts the HTTP trigger server
// This enables manual backup triggers via HTTP endpoint
func StartTriggerServer(cfg Config, ctxPool *pool.ContextPool) error {
	if !cfg.HTTPTrigger.Enable {
		log.Info().Msg("HTTP trigger server is disabled")
		return nil
	}

	log.Info().Msgf("Starting HTTP trigger server on port %s", cfg.HTTPTrigger.Port)

	// Create HTTP multiplexer with routes
	mux := http.NewServeMux()

	// POST /trigger - trigger a manual backup
	mux.HandleFunc(cfg.HTTPTrigger.Path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Info().Msg("Received manual backup trigger request")

		// Trigger the backup in a goroutine (non-blocking)
		go func() {
			now := time.Now()
			if err := start(cfg, now); err != nil {
				log.Err(err).Msg("Manual backup failed")
				return
			}
			log.Info().Msg("Manual backup completed successfully")
		}()

		// Return immediately while backup runs in background
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TriggerResponse{
			Status:  "success",
			Message: "Backup trigger initiated",
		})
	})

	// GET /health - health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TriggerResponse{
			Status:  "healthy",
			Message: "HTTP trigger server is running",
		})
	})

	// Initialize HTTP server with timeouts
	triggerServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPTrigger.Port),
		Handler:      mux,
		ReadTimeout:  defaultTriggerReadTimeout,
		WriteTimeout: defaultTriggerWriteTimeout,
		IdleTimeout:  defaultTriggerIdleTimeout,
	}

	// Store cancel function for graceful shutdown
	_, triggerCancel = context.WithCancel(context.Background())

	// Start server in background with graceful shutdown
	ctxPool.Go(func(ctx context.Context) error {
		go func() {
			if err := triggerServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("HTTP trigger server ListenAndServe failed")
			}
		}()

		<-ctx.Done()
		log.Debug().Msg("Received shutdown signal for HTTP trigger server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), triggerShutdownTimeout)
		defer cancel()

		if err := triggerServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP trigger server Shutdown failed: %w", err)
		}

		log.Info().Msg("HTTP trigger server gracefully stopped")
		return nil
	})

	log.Info().Msgf("HTTP trigger server started at http://localhost:%s%s", cfg.HTTPTrigger.Port, cfg.HTTPTrigger.Path)
	return nil
}
