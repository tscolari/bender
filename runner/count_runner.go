package runner

import (
	"os/exec"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner"
	"code.cloudfoundry.org/commandrunner/linux_command_runner"
)

type CountRunner struct {
	cmdRunner commandrunner.CommandRunner
	counter   int
	args      []string
}

func NewCountRunner(counter int, command ...string) *CountRunner {
	cmdRunner := linux_command_runner.New()
	return NewCountRunnerWithCmdRunner(cmdRunner, counter, command...)
}

func NewCountRunnerWithCmdRunner(cmdRunner commandrunner.CommandRunner, counter int, command ...string) *CountRunner {
	return &CountRunner{
		cmdRunner: cmdRunner,
		counter:   counter,
		args:      command,
	}
}

func (r *CountRunner) Run(concurrency int, cancel chan bool) (Summary, error) {
	var summary Summary
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
	}
	return summary, nil
}

func (r *CountRunner) run() RunStats {
	var runStats RunStats
	runStats.StartTime = time.Now()
	cmd := exec.Command(r.args[0], r.args[1:]...)
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
