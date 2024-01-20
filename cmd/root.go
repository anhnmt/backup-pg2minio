package cmd

import (
	"time"

	"github.com/rs/zerolog/log"
)

func init() {
	bootstrap()
}

func Execute() {
	cfg, err := New()
	if err != nil {
		log.Panic().Err(err).Msg("Failed to init config")
	}

	if cfg.Telegram.Enable {
		t, err := NewTelegram(cfg.Telegram, cfg.Postgres.Database)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to init telegram")
		}

		SetDefault(t)
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
}
