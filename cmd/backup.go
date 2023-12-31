package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func schedule(schedule string) {
	if viper.GetBool(TelegramEnabled) {
		t, err := NewTelegram()
		if err != nil {
			log.Err(err).Msg("Failed to init telegram")
			return
		}

		SetDefault(t)
	}

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

		if err := start(); err != nil {
			Err(err, "Failed to start backup")
		}

		OK("Backup successful: %s", time.Since(now).String())
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

func start() error {
	defer func() {
		if err := removeFile(PgDumpFile); err != nil {
			Err(err, "Failed to remove pg_dump file")
		}
	}()

	err := pgDump()
	if err != nil {
		return err
	}

	err = storage()
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
	case <-time.After(time.Millisecond):
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
