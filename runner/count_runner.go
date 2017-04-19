package runner

import (
	"os/exec"
	"time"

	"code.cloudfoundry.org/commandrunner"
)

type CountRunner struct {
	cmdRunner commandrunner.CommandRunner
	counter   int
	args      []string
}

func NewCountRunner(cmdRunner commandrunner.CommandRunner, counter int, command ...string) *CountRunner {
	return &CountRunner{
		cmdRunner: cmdRunner,
		counter:   counter,
		args:      command,
	}
}

func (r *CountRunner) Run() (Summary, error) {
	var summary Summary
	start := time.Now()

	for i := 0; i < r.counter; i++ {
		cmd := exec.Command(r.args[0], r.args[1:]...)
		err := r.cmdRunner.Run(cmd)
		if err != nil {
			summary.ErrorCounter++
		} else {
			summary.SuccessCounter++
		}
	}

	summary.Duration = time.Since(start)
	return summary, nil
}
