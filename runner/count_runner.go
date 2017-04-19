package runner

import (
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner"
	"code.cloudfoundry.org/commandrunner/linux_command_runner"
)

type CountRunner struct {
	cmdRunner commandrunner.CommandRunner
	counter   int
	commands  []string
}

func NewCountRunner(counter int, commands ...string) *CountRunner {
	cmdRunner := linux_command_runner.New()
	return NewCountRunnerWithCmdRunner(cmdRunner, counter, commands...)
}

func NewCountRunnerWithCmdRunner(cmdRunner commandrunner.CommandRunner, counter int, commands ...string) *CountRunner {
	return &CountRunner{
		cmdRunner: cmdRunner,
		counter:   counter,
		commands:  commands,
	}
}

func (r *CountRunner) Run(concurrency int, cancel chan bool) (Summary, error) {
	summary := Summary{
		Commands: r.commandsSummary(),
	}

	start := time.Now()
	wg := sync.WaitGroup{}

	tasks := make(chan bool, r.counter)
	stats := make(chan RunStats, r.counter)

	for i := 0; i < r.counter; i++ {
		tasks <- true
	}

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			r.startWorker(tasks, cancel, stats)
			wg.Done()
		}()
	}

	wg.Wait()
	summary.Duration = time.Since(start)

	close(stats)
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
	return summary, nil
}

func (r *CountRunner) run() RunStats {
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

func (r *CountRunner) startWorker(tasks chan bool, stop chan bool, stats chan RunStats) {
	for {
		select {
		case <-tasks:
			stats <- r.run()
		case <-stop:
			return
		default:
			return
		}
	}
}

func (r *CountRunner) commandsSummary() map[int]Command {
	summary := map[int]Command{}

	for i, command := range r.commands {
		summary[i+1] = Command{
			Exec: command,
		}
	}

	return summary
}
