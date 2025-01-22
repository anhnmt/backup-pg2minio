package cmd

import (
	"context"
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
	defaultSessionTimeout = 6 * time.Second
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctxPool := pool.New().
		WithContext(ctx).
		WithFirstError().
		WithCancelOnError()

	if cfg.Metrics.Enable && cfg.Schedule.Cron != "" {
		mux := http.NewServeMux()

		promMetrics, promReg, err := NewPrometheusMetrics(cfg.Metrics.Namespace, cfg.Metrics.Subsystem)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to init prometheus metrics")
		}

		mux.Handle(cfg.Metrics.Path, promhttp.HandlerFor(promReg, promhttp.HandlerOpts{}))

		SetDefaultMetrics((promMetrics))

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

	err = aliasSet(cfg.Minio)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to set alias minio")
	}

	if cfg.Minio.Prerun {
		if err = preRunMinio(cfg.Minio); err != nil {
			log.Panic().Err(err).Msg("Failed to pre-run minio")
		}
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
		log.Panic().Err(err).Msg("error in goroutine pool")
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

	log.Info().Msg("initialized metrics HTTP server")

	ctxPool.Go(func(ctx context.Context) error {
		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("HTTP server ListenAndServe failed")
			}
		}()

		<-ctx.Done()
		log.Debug().Msg("Received shutdown signal, initiating graceful shutdown")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("HTTP server Shutdown failed")
			return fmt.Errorf("HTTP server Shutdown failed: %w", err)
		}

		log.Info().Msg("HTTP server gracefully stopped")
		return nil
	})
}
