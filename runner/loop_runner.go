package runner

import (
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner"
	"code.cloudfoundry.org/commandrunner/linux_command_runner"
)

type LoopRunner struct {
	baseRunner
	interval time.Duration
}

func NewLoopRunner(interval time.Duration, commands ...string) *LoopRunner {
	cmdRunner := linux_command_runner.New()
	return NewLoopRunnerWithCmdRunner(cmdRunner, interval, commands...)
}

func NewLoopRunnerWithCmdRunner(cmdRunner commandrunner.CommandRunner, interval time.Duration, commands ...string) *LoopRunner {
	baseRunner := newBaseRunner(cmdRunner, commands...)

	return &LoopRunner{
		baseRunner: baseRunner,
		interval:   interval,
	}
}

func (r *LoopRunner) Run(concurrency int, cancel chan bool) (Summary, error) {
	summary := Summary{
		Commands: r.commandsSummary(),
	}

	start := time.Now()
	wg := sync.WaitGroup{}

	stats := make(chan RunStats, 1000)

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			r.startWorker(r.interval, cancel, stats)
			wg.Done()
		}()
	}

	mergeStatsDone := make(chan bool)
	go func() {
		r.mergeRunstatsIntoSummary(stats, &summary)
		close(mergeStatsDone)
	}()

	wg.Wait()
	close(stats)
	<-mergeStatsDone

	summary.Duration = time.Since(start)
	return summary, nil
}

func (r *LoopRunner) startWorker(interval time.Duration, stop chan bool, stats chan RunStats) {
	for {
		select {
		case <-stop:
			return
		default:
			stats <- r.run()

			select {
			case <-stop:
				return
			case <-time.After(interval):
				break
			}
		}
	}
}
