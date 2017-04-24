package runner

import (
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"code.cloudfoundry.org/commandrunner"
)

// Defines the summary of a Runner.Run
// - Commands is the a indexed map of all the commands that were ran
// - Duration is the total duration of the Run method
// - SuccessCounter totalizes the total of times the commands were ran with success
// - ErrorCounter totalizes the total of tiems the commands were ran with failure
// - EachRun contains the information of each ran of the commands
type Summary struct {
	Commands       map[int]Command `json:"commands"`
	Duration       time.Duration   `json:"duration"`
	SuccessCounter int             `json:"success_counter"`
	ErrorCounter   int             `json:"error_counter"`
	EachRun        []RunStats      `json:"each_run"`
}

// Contains information about each of the times the commands were executed
// - Command is the index of the command (from Summary.Commands)
// - Duration is the duration of the command execution in this run
// - StartTime defines when this run started
// - Failed signilizes if the command returned any kind of error
type RunStats struct {
	Command   int           `json:"command"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	Failed    bool          `json:"failed"`
}

// Simple command information
// - Exec is the full command+args that were executed
// - RunCount is the total times this particular command were executed
type Command struct {
	Exec     string `json:"exec"`
	RunCount int    `json:"run_count"`
}

// Runner defines the interface for benchmarking a set of commands
// The Run method must take a concurrency level and a channel to allow cancelation.
// It also takes a list of commands to be ran, and return a Summary of the execution.
// error is intended only to be returned if there's an error with the setup,
// no with the command execution itself.
type Runner interface {
	Run(concurrency int, cancel chan bool, commands ...string) (Summary, error)
}

type baseRunner struct {
	cmdRunner commandrunner.CommandRunner
}

func newBaseRunner(cmdRunner commandrunner.CommandRunner) baseRunner {
	return baseRunner{
		cmdRunner: cmdRunner,
	}
}

func (r *baseRunner) run(commands []string) RunStats {
	var runStats RunStats
	runStats.StartTime = time.Now()

	cmdIdx := rand.Int() % len(commands)
	args := strings.Split(commands[cmdIdx], " ")
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

func (r *baseRunner) commandsSummary(commands []string) map[int]Command {
	summary := map[int]Command{}

	for i, command := range commands {
		summary[i+1] = Command{
			Exec: command,
		}
	}

	return summary
}
