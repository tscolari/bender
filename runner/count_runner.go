package runner

import (
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner"
	"code.cloudfoundry.org/commandrunner/linux_command_runner"
)

type CountRunner struct {
	baseRunner
	counter int
}

func NewCountRunner(counter int, commands ...string) *CountRunner {
	cmdRunner := linux_command_runner.New()
	return NewCountRunnerWithCmdRunner(cmdRunner, counter, commands...)
}

func NewCountRunnerWithCmdRunner(cmdRunner commandrunner.CommandRunner, counter int, commands ...string) *CountRunner {
	baseRunner := newBaseRunner(cmdRunner, commands...)
	return &CountRunner{
		baseRunner: baseRunner,
		counter:    counter,
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
	r.mergeRunstatsIntoSummary(stats, &summary)
	return summary, nil
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
