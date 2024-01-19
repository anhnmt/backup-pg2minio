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
		return
	}

	if cfg.Telegram.Enable {
		t, err := NewTelegram(cfg.Telegram, cfg.Postgres.Database)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to init telegram")
			return
		}

		SetDefault(t)
	}

	if cfg.Schedule.Cron == "" {
		log.Info().Msgf("Start backup")

		if err = start(cfg, time.Now()); err != nil {
			log.Panic().Err(err).Msg("Failed to start backup")
			return
		}

		log.Info().Msg("Backup successfully")
		return
	}

	Cron(cfg)
}
