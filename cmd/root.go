package cmd

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	bootstrap()
}

func Execute() {
	backupSchedule := viper.GetString(BackupSchedule)

	if backupSchedule == "" {
		log.Info().Msgf("Start backup")

		if err := start(time.Now()); err != nil {
			log.Panic().Err(err).Msg("Failed to start backup")
			return
		}

		log.Info().Msg("Backup successfully")
		return
	}

	schedule(backupSchedule)
}
