package main

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/anhnmt/backup-pg2minio/internal/pkg/config"
	"github.com/anhnmt/backup-pg2minio/internal/pkg/minio"
	"github.com/anhnmt/backup-pg2minio/internal/pkg/postgres"
	"github.com/anhnmt/backup-pg2minio/internal/pkg/telegram"
)

func init() {
	bootstrap()
}

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Panic().Err(err).Msg("Failed to init config")
	}

	if cfg.Telegram.Enable {
		t, err := telegram.NewTelegram(cfg.Telegram, cfg.Postgres.Database)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to init telegram")
		}

		telegram.SetDefault(t)
	}

	if cfg.Postgres.Prerun {
		if err = postgres.PreRunPostgres(cfg.Postgres); err != nil {
			log.Panic().Err(err).Msg("Failed to pre-run postgres")
		}
	}

	err = minio.AliasSet(cfg.Minio)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to set alias minio")
	}

	if cfg.Minio.Prerun {
		if err = minio.PreRunMinio(cfg.Minio); err != nil {
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
}
