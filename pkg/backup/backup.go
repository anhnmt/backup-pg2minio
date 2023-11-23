package backup

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

func Schedule(schedule string) {
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

		// err := utils.Execute(command, args)
		// if err != nil {
		// 	log.Err(err).Msg("Failed to execute command")
		// }
	})
	if err != nil {
		log.Panic().Err(err).Msg("Failed to add cron job")
		return
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	// stop(c, wg)
	return
}

func Start() error {

	return nil
}
