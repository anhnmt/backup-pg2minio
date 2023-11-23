package main

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/robfig/cron"
)

func execute(command string, args []string) {

	println("executing:", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	cmd.Wait()
}

func create() (cr *cron.Cron, wgr *sync.WaitGroup) {
	var schedule = os.Args[1]
	var command = os.Args[2]
	var args = os.Args[3:len(os.Args)]

	wg := &sync.WaitGroup{}

	c := cron.New()
	println("new cron:", schedule)

	c.AddFunc(schedule, func() {
		wg.Add(1)
		execute(command, args)
		wg.Done()
	})

	return c, wg
}

func start(c *cron.Cron) {
	c.Start()
}

func stop(c *cron.Cron, wg *sync.WaitGroup) {
	println("Stopping")
	c.Stop()
	println("Waiting")
	wg.Wait()
	println("Exiting")
	os.Exit(0)
}

func main() {

	c, wg := create()

	go start(c)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	println(<-ch)

	stop(c, wg)
}
