package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tscolari/bender/runner"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "bender"
	app.Version = "1.0.0"
	app.Usage = "Benchmark external commands"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "count",
			Value: 1,
			Usage: "how many times should the command run",
		},
		cli.IntFlag{
			Name:  "concurrency",
			Value: 1,
			Usage: "how many threads to use",
		},
		cli.StringSliceFlag{
			Name:  "command",
			Usage: "command(s) to run. May be set more than once",
		},
		cli.BoolFlag{
			Name:  "keep-running",
			Usage: "run until aborted (ctrl-c)",
		},
		cli.DurationFlag{
			Name:  "interval",
			Value: 0,
			Usage: "interval to use between each call when using keep-running",
		},
	}

	app.Action = func(c *cli.Context) error {
		if len(c.StringSlice("command")) == 0 {
			return errors.New("Missing at least one `--command` argument")
		}

		if c.Bool("keep-running") && c.IsSet("count") {
			return errors.New("can't use `--keep-running` and `--count` at the same time")
		}

		if c.IsSet("count") && c.IsSet("interval") {
			return errors.New("can't use `--count` and `--interval` at the same time")
		}

		cancelChan := make(chan bool)
		listenForShutdown(cancelChan)

		runner, err := newRunnerFromArgs(c)
		if err != nil {
			return err
		}

		summary, err := runner.Run(c.Int("concurrency"), cancelChan, c.StringSlice("command")...)
		if err != nil {
			return fmt.Errorf("Failed to run: %s", err.Error())
		}

		err = json.NewEncoder(os.Stdout).Encode(&summary)
		if err != nil {
			return fmt.Errorf("Failed to run: %s", err.Error())
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

}

func newRunnerFromArgs(c *cli.Context) (runner.Runner, error) {
	if c.IsSet("count") {
		return runner.NewCountRunner(c.Int("count")), nil
	}

	if c.IsSet("keep-running") {
		return runner.NewLoopRunner(c.Duration("interval")), nil
	}

	return nil, errors.New("no runner detected. Use `--keep-running` or `--count`")
}

func listenForShutdown(cancel chan bool) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt, syscall.SIGKILL)
	go func() {
		<-c
		close(cancel)
	}()
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
