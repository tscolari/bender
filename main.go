package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/tscolari/bender/runner"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "bender"
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
	}

	app.Action = func(c *cli.Context) error {
		commands := c.StringSlice("command")
		if len(commands) == 0 {
			return errors.New("Missing at least one `--command` argument")
		}

		runner := runner.NewCountRunner(c.Int("count"), commands...)
		cancelChan := make(chan bool)
		summary, err := runner.Run(c.Int("concurrency"), cancelChan)
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

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
