package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/docker/go-units"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/anhnmt/backup-pg2minio/internal/pkg/config"
	"github.com/anhnmt/backup-pg2minio/internal/pkg/minio"
	"github.com/anhnmt/backup-pg2minio/internal/pkg/postgres"
	"github.com/anhnmt/backup-pg2minio/internal/pkg/telegram"
	"github.com/anhnmt/backup-pg2minio/internal/utils"
)

func Cron(cfg config.Config) {
	schedule := cfg.Schedule.Cron

	log.Info().Msgf("New cron: %s", schedule)

	var opts []cron.Option
	if len(strings.Split(schedule, " ")) >= 6 {
		opts = append(opts, cron.WithSeconds())
	}

	wg := &sync.WaitGroup{}
	c := cron.New(opts...)

	_, err := c.AddFunc(schedule, func() {
		wg.Add(1)
		defer wg.Done()

		now := time.Now()
		log.Info().Msgf("Start backup at: %s", now.Format(time.RFC3339))

		if err := start(cfg, now); err != nil {
			telegram.Err(err, "Failed to start backup")
		}
	})
	if err != nil {
		log.Panic().Err(err).Msg("Failed to add cron job")
		return
	}

	c.Start()
	log.Info().Msg("Cron start")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	stop(c, wg)
	return
}

func start(cfg config.Config, now time.Time) (err error) {
	defer func(_err *error) {
		info, err2 := os.Stat(utils.PgDumpFile)
		if os.IsNotExist(err2) {
			return
		}

		if err2 = os.Remove(utils.PgDumpFile); err2 != nil {
			telegram.Err(err, "Failed to remove pg_dump file")
		}

		if *_err == nil {
			telegram.OK("Backup successful: %s, size: %s",
				time.Since(now).String(),
				units.BytesSize(float64(info.Size())),
			)
		}

	}(&err)

	err = postgres.PgDump(cfg.Postgres)
	if err != nil {
		return err
	}

	err = minio.Storage(cfg.Minio, cfg.Postgres.Database)
	if err != nil {
		return err
	}

	return nil
}

func stop(c *cron.Cron, wg *sync.WaitGroup) {
	log.Info().Msg("Stopping")
	ctx := c.Stop()
	select {
	case <-ctx.Done():
		// expected
	case <-time.After(5 * time.Second):
		log.Panic().
			Err(fmt.Errorf("context not done even when cron Stop is completed")).
			Msg("Failed to stop cron")
		return
	}

	log.Info().Msg("Waiting")
	wg.Wait()

	log.Info().Msg("Exiting")
	os.Exit(0)
}
