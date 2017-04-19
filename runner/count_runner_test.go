package runner_test

import (
	"errors"
	"os/exec"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner/fake_command_runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tscolari/bender/runner"
)

var _ = Describe("CountRunner", func() {
	var (
		args        []string
		cancelChan  chan bool
		count       int
		cmdRunner   *fake_command_runner.FakeCommandRunner
		countRunner *runner.CountRunner

		commandFunc func(cmd *exec.Cmd) error
	)

	BeforeEach(func() {
		cancelChan = make(chan bool)
		commandFunc = nil
		cmdRunner = fake_command_runner.New()
		args = []string{"hello", "world"}
		count = 6

		cmdRunner.WhenRunning(fake_command_runner.CommandSpec{
			Path: args[0],
			Args: args[1:],
		}, func(cmd *exec.Cmd) error {
			if commandFunc != nil {
				return commandFunc(cmd)
			}

			return nil
		})
	})

	JustBeforeEach(func() {
		countRunner = runner.NewCountRunnerWithCmdRunner(cmdRunner, count, args...)
	})

	Describe("Run", func() {
		It("calls the command runner with correct arguments", func() {
			_, err := countRunner.Run(1, cancelChan)
			Expect(err).NotTo(HaveOccurred())

			executedCommands := cmdRunner.ExecutedCommands()
			Expect(executedCommands).To(HaveLen(count))

			for _, cmd := range executedCommands {
				Expect(cmd.Args).To(Equal(args))
			}
		})

		It("summarizes the total duration", func() {
			commandFunc = func(_ *exec.Cmd) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			}

			summary, err := countRunner.Run(1, cancelChan)
			Expect(err).NotTo(HaveOccurred())

			Expect(summary.Duration).To(BeNumerically("~", time.Duration(count)*(10*time.Millisecond), 3*time.Millisecond))
		})

		It("summarizes the success/failures", func() {
			mutex := &sync.Mutex{}
			i := 0

			commandFunc = func(_ *exec.Cmd) error {
				mutex.Lock()
				i++
				mutex.Unlock()
				if i%2 == 0 {
					return errors.New("not odd!")
				}
				return nil
			}

			summary, err := countRunner.Run(1, cancelChan)
			Expect(err).NotTo(HaveOccurred())

			Expect(summary.SuccessCounter).To(Equal(count / 2))
			Expect(summary.ErrorCounter).To(Equal(count / 2))
		})

		It("summarizes run details for each run", func() {
			mutex := &sync.Mutex{}
			i := 0

			commandFunc = func(_ *exec.Cmd) error {
				time.Sleep(10 * time.Millisecond)
				mutex.Lock()
				i++
				mutex.Unlock()
				if i%2 == 0 {
					return errors.New("odd!")
				}
				return nil
			}

			start := time.Now()
			summary, err := countRunner.Run(1, cancelChan)
			Expect(err).NotTo(HaveOccurred())

			Expect(summary.EachRun).To(HaveLen(6))
			Expect(summary.EachRun[0].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
			Expect(summary.EachRun[0].Failed).To(BeFalse())
			Expect(summary.EachRun[0].StartTime.UnixNano()).To(BeNumerically(">", start.UnixNano()))

			Expect(summary.EachRun[1].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
			Expect(summary.EachRun[1].Failed).To(BeTrue())
			Expect(summary.EachRun[1].StartTime.UnixNano()).To(BeNumerically(">", summary.EachRun[0].StartTime.UnixNano()))

			Expect(summary.EachRun[2].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
			Expect(summary.EachRun[2].Failed).To(BeFalse())
			Expect(summary.EachRun[2].StartTime.UnixNano()).To(BeNumerically(">", summary.EachRun[1].StartTime.UnixNano()))

			Expect(summary.EachRun[3].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
			Expect(summary.EachRun[3].Failed).To(BeTrue())
			Expect(summary.EachRun[3].StartTime.UnixNano()).To(BeNumerically(">", summary.EachRun[2].StartTime.UnixNano()))

			Expect(summary.EachRun[4].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
			Expect(summary.EachRun[4].Failed).To(BeFalse())
			Expect(summary.EachRun[4].StartTime.UnixNano()).To(BeNumerically(">", summary.EachRun[3].StartTime.UnixNano()))
		})

		Context("cancelling", func() {
			It("is possible to cancel a running job", func() {
				commandFunc = func(_ *exec.Cmd) error {
					time.Sleep(1 * time.Second)
					return nil
				}

				finished := false
				mutex := &sync.Mutex{}

				go func() {
					_, err := countRunner.Run(1, cancelChan)
					Expect(err).NotTo(HaveOccurred())
					mutex.Lock()
					finished = true
					mutex.Unlock()
				}()

				close(cancelChan)

				Eventually(func() bool {
					mutex.Lock()
					defer mutex.Unlock()
					return finished
				}, 1*time.Second).Should(BeTrue())

			})
		})

		Context("concurrency", func() {
			It("runs commands in parallel", func() {
				commandFunc = func(_ *exec.Cmd) error {
					time.Sleep(10 * time.Millisecond)
					return nil
				}

				start := time.Now()
				summary, err := countRunner.Run(6, cancelChan)
				Expect(err).NotTo(HaveOccurred())

				Expect(summary.EachRun).To(HaveLen(6))
				Expect(summary.EachRun[0].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
				Expect(summary.EachRun[0].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))

				Expect(summary.EachRun[1].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
				Expect(summary.EachRun[1].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))

				Expect(summary.EachRun[2].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
				Expect(summary.EachRun[2].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))

				Expect(summary.EachRun[3].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
				Expect(summary.EachRun[3].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))

				Expect(summary.EachRun[4].Duration).To(BeNumerically("~", 10*time.Millisecond, time.Millisecond))
				Expect(summary.EachRun[4].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))
			})
		})
	})
})
