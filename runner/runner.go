package runner

import (
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"code.cloudfoundry.org/commandrunner"
)

type Summary struct {
	Commands       map[int]Command `json:"commands"`
	Duration       time.Duration   `json:"duration"`
	SuccessCounter int             `json:"success_counter"`
	ErrorCounter   int             `json:"error_counter"`
	EachRun        []RunStats      `json:"each_run"`
}

type RunStats struct {
	Command   int           `json:"command"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	Failed    bool          `json:"failed"`
}

type Command struct {
	Exec     string `json:"exec"`
	RunCount int    `json:"run_count"`
}

type Runner interface {
	Run(concurrency int, cancel chan bool) (Summary, error)
}

type baseRunner struct {
	commands  []string
	cmdRunner commandrunner.CommandRunner
}

func newBaseRunner(cmdRunner commandrunner.CommandRunner, commands ...string) baseRunner {
	return baseRunner{
		commands:  commands,
		cmdRunner: cmdRunner,
	}
}

func (r *baseRunner) run() RunStats {
	var runStats RunStats
	runStats.StartTime = time.Now()

	cmdIdx := rand.Int() % len(r.commands)
	args := strings.Split(r.commands[cmdIdx], " ")
	runStats.Command = cmdIdx + 1

	cmd := exec.Command(args[0], args[1:]...)
	err := r.cmdRunner.Run(cmd)
	runStats.Duration = time.Since(runStats.StartTime)
	if err != nil {
		runStats.Failed = true
	}

	return runStats
}

func (r *baseRunner) mergeRunstatsIntoSummary(stats chan RunStats, summary *Summary) {
	for runStats := range stats {
		if runStats.Failed {
			summary.ErrorCounter++
		} else {
			summary.SuccessCounter++
		}
		summary.EachRun = append(summary.EachRun, runStats)

		cmd := summary.Commands[runStats.Command]
		cmd.RunCount++
		summary.Commands[runStats.Command] = cmd
	}
}

func (r *baseRunner) commandsSummary() map[int]Command {
	summary := map[int]Command{}

	for i, command := range r.commands {
		summary[i+1] = Command{
			Exec: command,
		}
	}

	return summary
}
