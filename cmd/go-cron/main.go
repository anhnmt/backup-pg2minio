package main

import (
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// UNIX Time is faster and smaller than most timestamps
	consoleWriter := &zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	// Caller Marshal Function
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	log.Logger = zerolog.
		New(consoleWriter).
		With().
		Timestamp().
		Caller().
		Logger()
}

func execute(command string, args []string) error {
	log.Info().Msgf("Executing: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func create() (cr *cron.Cron, wgr *sync.WaitGroup) {
	var schedule = os.Args[1]
	var command = os.Args[2]
	var args = os.Args[3:len(os.Args)]

	wg := &sync.WaitGroup{}
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

	return c, wg
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
	c, wg := create()
	go start(c)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	stop(c, wg)
}
