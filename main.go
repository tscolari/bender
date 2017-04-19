package main

import (
	"encoding/json"
	"fmt"
	"os"

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
	}

	app.Action = func(c *cli.Context) error {
		runner := runner.NewCountRunner(c.Int("count"), c.Args()...)
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
