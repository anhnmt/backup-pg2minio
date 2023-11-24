package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	bootstrap()
}

func Execute() {
	schedule := viper.GetString("SCHEDULE")

	if schedule == "" {
		log.Info().Msgf("Start backup")

		if err := start(); err != nil {
			log.Panic().Err(err).Msg("Failed to start backup")
			return
		}

		log.Info().Msg("Backup successfully")
		return
	}

	scheduleBackup(schedule)
}
