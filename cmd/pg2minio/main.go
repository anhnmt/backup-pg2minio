package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/anhnmt/backup-pg2minio/pkg/backup"
	"github.com/anhnmt/backup-pg2minio/pkg/bootstrap"
)

func init() {
	bootstrap.Bootstrap()
}

func main() {
	schedule := viper.GetString("SCHEDULE")

	if schedule == "" {
		log.Info().Msgf("Start backup")

		err := backup.Start()
		if err != nil {
			log.Panic().Err(err).Msg("Failed to start backup")
			return
		}

		log.Info().Msg("Backup successfully")
		return
	}

	backup.Schedule(schedule)
}
