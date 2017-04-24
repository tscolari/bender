package runner

import (
	"errors"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner"
	"code.cloudfoundry.org/commandrunner/linux_command_runner"
)

// LoopRunner defines a runner that runs until it gets canceled.
// To cancel it's execution the `cancel` channel in the `Run` method must be closed.
// It's possible to define a interval in which the LoopRunner will wait to execute
// the next command. Note that if concurrency level is anything bigger than 1,
// interval will apply for each thread individually.
type LoopRunner struct {
	baseRunner
	interval time.Duration
}

// Creates a new instance of the LoopRunner.
// To run without any interval between runs, set interval to 0
func NewLoopRunner(interval time.Duration) *LoopRunner {
	cmdRunner := linux_command_runner.New()
	return NewLoopRunnerWithCmdRunner(cmdRunner, interval)
}

func NewLoopRunnerWithCmdRunner(cmdRunner commandrunner.CommandRunner, interval time.Duration) *LoopRunner {
	baseRunner := newBaseRunner(cmdRunner)

	return &LoopRunner{
		baseRunner: baseRunner,
		interval:   interval,
	}
}

// Start commands execution.
// This method will block until `cancel` is closed. Once cancel is closed, it will
// wait for the any running command to finish and summarize the results.
func (r *LoopRunner) Run(concurrency int, cancel chan bool, commands ...string) (Summary, error) {
	if len(commands) == 0 {
		return Summary{}, errors.New("no commands given")
	}

	summary := Summary{
		Commands: r.commandsSummary(commands),
	}

	start := time.Now()
	wg := sync.WaitGroup{}

	stats := make(chan RunStats, 1000)

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			r.startWorker(r.interval, cancel, stats, commands)
			wg.Done()
		}()
	}

	mergeStatsDone := make(chan bool)
	go func() {
		r.mergeRunstatsIntoSummary(stats, &summary)
		close(mergeStatsDone)
	}()

	wg.Wait()
	summary.Duration = time.Since(start)
	close(stats)
	<-mergeStatsDone
	return summary, nil
}

func (r *LoopRunner) startWorker(interval time.Duration, stop chan bool, stats chan RunStats, commands []string) {
	for {
		select {
		case <-stop:
			return
		default:
			stats <- r.run(commands)

			select {
			case <-stop:
				return
			case <-time.After(interval):
				break
			}
		}
	}
}
