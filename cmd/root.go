package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"
)

const (
	defaultReadTimeout    = 5 * time.Second
	defaultWriteTimeout   = 10 * time.Second
	defaultIdleTimeout    = 30 * time.Second
	serverShutdownTimeout = 10 * time.Second
)

func init() {
	bootstrap()
}

func Execute() {
	cfg, err := New()

	if err != nil {
		log.Panic().Err(err).Msg("Failed to init config")
	}

	// Determine action: backup or restore
	switch cfg.Action {
	case ActionRestore:
		// Restore mode
		if cfg.Restore.SourcePath == "" {
			log.Panic().Msg("RESTORE_SOURCE_PATH is required for restore action")
		}

		restoreCfg := RestoreConfig{
			Postgres:   cfg.Postgres,
			Minio:      cfg.Minio,
			SourcePath: cfg.Restore.SourcePath,
			TargetDB:   cfg.Restore.TargetDB,
		}

		if err := PerformRestore(restoreCfg); err != nil {
			log.Panic().Err(err).Msg("Restore failed")
		}

		log.Info().Msg("Restore completed successfully")
		return

	case ActionBackup:
		// Backup mode - continue with normal flow
		log.Info().Msg("Running in backup mode")

	default:
		log.Panic().Msgf("Invalid ACTION: %s. Must be 'backup' or 'restore'", cfg.Action)
	}

	if cfg.Metrics.Enable && cfg.Schedule.Cron == "" {
		log.Panic().Msg("Metrics cannot be enabled without a cron schedule")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctxPool := pool.New().
		WithContext(ctx).
		WithFirstError().
		WithCancelOnError()

	if cfg.HTTPTrigger.Enable {
		if err := StartTriggerServer(cfg, ctxPool); err != nil {
			log.Panic().Err(err).Msg("Failed to start trigger server")
		}
	}

	if cfg.Metrics.Enable {
		mux := http.NewServeMux()

		promMetrics, promReg, err := NewPrometheusMetrics(cfg.Metrics.Namespace, cfg.Metrics.Subsystem)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to init prometheus metrics")
		}

		mux.Handle(cfg.Metrics.Path, promhttp.HandlerFor(promReg, promhttp.HandlerOpts{}))

		SetDefaultMetrics(promMetrics)

		setupHTTPServer(ctxPool, mux, cfg)
	}

	if cfg.Telegram.Enable {
		t, err := NewTelegram(cfg.Telegram, cfg.Postgres.Database)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to init telegram")
		}

		SetDefaultTelegram(t)
	}

	if cfg.Postgres.Prerun {
		if err = preRunPostgres(cfg.Postgres); err != nil {
			log.Panic().Err(err).Msg("Failed to pre-run postgres")
		}
	}

	// Note: aliasSet and preRunMinio are no longer needed with MinIO SDK
	// The MinIO client is created directly in storage() and restore functions

	if cfg.Minio.Prerun {
		// MinIO bucket check is now done in storage() function
		log.Info().Msg("MinIO prerun check skipped - will be validated during upload")
	}

	if cfg.Schedule.Cron == "" {
		log.Info().Msgf("Start backup")

		if err = start(cfg, time.Now()); err != nil {
			log.Panic().Err(err).Msg("Failed to start backup")
		}

		log.Info().Msg("Backup successfully")
		return
	}

	Cron(cfg)

	if err = ctxPool.Wait(); err != nil {
		log.Panic().Err(err).Msg("Error in goroutine pool")
	}
}

func setupHTTPServer(ctxPool *pool.ContextPool, mux *http.ServeMux, env Config) {
	httpServer := http.Server{
		Addr:         fmt.Sprint(":", env.Metrics.Port),
		Handler:      mux,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	log.Info().Msg("Initialized metrics HTTP server")

	ctxPool.Go(func(ctx context.Context) error {
		go func() {
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("HTTP server ListenAndServe failed")
			}
		}()

		<-ctx.Done()
		log.Debug().Msg("Received shutdown signal, initiating graceful shutdown")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server Shutdown failed: %w", err)
		}

		log.Info().Msg("HTTP server gracefully stopped")
		return nil
	})
}
