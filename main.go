package main

import (
	"encoding/json"
	"fmt"
	"os"

	"code.cloudfoundry.org/commandrunner/linux_command_runner"

	"github.com/tscolari/bender/runner"
)

func main() {
	runner := runner.NewCountRunner(&linux_command_runner.RealCommandRunner{}, 3, os.Args[3], os.Args[4])
	summary, err := runner.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %s", err.Error())
		os.Exit(1)
	}

	err = json.NewEncoder(os.Stdout).Encode(&summary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run: %s", err.Error())
		os.Exit(1)
	}
}
