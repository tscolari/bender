package runner

import (
	"errors"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner"
	"code.cloudfoundry.org/commandrunner/linux_command_runner"
)

type CountRunner struct {
	baseRunner
	counter int
}

func NewCountRunner(counter int) *CountRunner {
	cmdRunner := linux_command_runner.New()
	return NewCountRunnerWithCmdRunner(cmdRunner, counter)
}

func NewCountRunnerWithCmdRunner(cmdRunner commandrunner.CommandRunner, counter int) *CountRunner {
	baseRunner := newBaseRunner(cmdRunner)
	return &CountRunner{
		baseRunner: baseRunner,
		counter:    counter,
	}
}

func (r *CountRunner) Run(concurrency int, cancel chan bool, commands ...string) (Summary, error) {
	if len(commands) == 0 {
		return Summary{}, errors.New("no commands given")
	}

	summary := Summary{
		Commands: r.commandsSummary(commands),
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
			r.startWorker(tasks, cancel, stats, commands)
			wg.Done()
		}()
	}

	wg.Wait()
	summary.Duration = time.Since(start)

	close(stats)
	r.mergeRunstatsIntoSummary(stats, &summary)
	return summary, nil
}

func (r *CountRunner) startWorker(tasks chan bool, stop chan bool, stats chan RunStats, commands []string) {
	for {
		select {
		case <-tasks:
			stats <- r.run(commands)
		case <-stop:
			return
		default:
			return
		}
	}
}
