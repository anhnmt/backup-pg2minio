package main

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/anhnmt/backup-pg2minio/pkg/bootstrap"
)

func init() {
	bootstrap.Bootstrap()
}

func execute(command string, args []string) error {
	log.Info().Msgf("Executing: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func create(wg *sync.WaitGroup) (cr *cron.Cron) {
	var schedule = os.Args[1]
	var command = os.Args[2]
	var args = os.Args[3:len(os.Args)]

	log.Info().Msgf("New cron: %s", schedule)

	var opts []cron.Option
	if len(strings.Split(schedule, " ")) >= 6 {
		opts = append(opts, cron.WithSeconds())
	}

	c := cron.New(opts...)

	_, err := c.AddFunc(schedule, func() {
		wg.Add(1)
		defer wg.Done()

		err := execute(command, args)
		if err != nil {
			log.Err(err).Msg("Failed to execute command")
		}
	})
	if err != nil {
		log.Panic().Err(err).Msg("Failed to add cron job")
	}

	return c
}

func start(c *cron.Cron) {
	c.Start()
}

func stop(c *cron.Cron, wg *sync.WaitGroup) {
	log.Info().Msg("Stopping")
	c.Stop()

	log.Info().Msg("Waiting")
	wg.Wait()

	log.Info().Msg("Exiting")
	os.Exit(0)
}

func main() {
	wg := &sync.WaitGroup{}

	c := create(wg)
	go start(c)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	stop(c, wg)
}
