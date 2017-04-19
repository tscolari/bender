package runner_test

import (
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner/fake_command_runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tscolari/bender/runner"
)

var _ = Describe("CountRunner", func() {
	var (
		commands    []string
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
		commands = []string{"hello world"}
		count = 6
	})

	JustBeforeEach(func() {
		countRunner = runner.NewCountRunnerWithCmdRunner(cmdRunner, count, commands...)

		cmdRunner.WhenRunning(fake_command_runner.CommandSpec{
			Path: strings.Split(commands[0], " ")[0],
			Args: strings.Split(commands[0], " ")[1:],
		}, func(cmd *exec.Cmd) error {
			if commandFunc != nil {
				return commandFunc(cmd)
			}

			return nil
		})
	})

	Describe("Run", func() {
		It("calls the command runner with correct arguments", func() {
			_, err := countRunner.Run(1, cancelChan)
			Expect(err).NotTo(HaveOccurred())

			executedCommands := cmdRunner.ExecutedCommands()
			Expect(executedCommands).To(HaveLen(count))

			for _, cmd := range executedCommands {
				Expect(cmd.Args).To(Equal(strings.Split(commands[0], " ")))
			}
		})

		It("summarizes the total duration", func() {
			commandFunc = func(_ *exec.Cmd) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			}

			summary, err := countRunner.Run(1, cancelChan)
			Expect(err).NotTo(HaveOccurred())

			Expect(summary.Duration).To(BeNumerically("~", time.Duration(count)*(10*time.Millisecond), 10*time.Millisecond))
		})

		It("summarizes the commands it ran", func() {
			summary, err := countRunner.Run(1, cancelChan)
			Expect(err).NotTo(HaveOccurred())
			Expect(summary.Commands).To(HaveLen(1))
			Expect(summary.Commands[1].Exec).To(Equal("hello world"))
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
			Expect(summary.EachRun[0].Command).To(Equal(1))
			Expect(summary.EachRun[0].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
			Expect(summary.EachRun[0].Failed).To(BeFalse())
			Expect(summary.EachRun[0].StartTime.UnixNano()).To(BeNumerically(">", start.UnixNano()))

			Expect(summary.EachRun[1].Command).To(Equal(1))
			Expect(summary.EachRun[1].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
			Expect(summary.EachRun[1].Failed).To(BeTrue())
			Expect(summary.EachRun[1].StartTime.UnixNano()).To(BeNumerically(">", summary.EachRun[0].StartTime.UnixNano()))

			Expect(summary.EachRun[2].Command).To(Equal(1))
			Expect(summary.EachRun[2].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
			Expect(summary.EachRun[2].Failed).To(BeFalse())
			Expect(summary.EachRun[2].StartTime.UnixNano()).To(BeNumerically(">", summary.EachRun[1].StartTime.UnixNano()))

			Expect(summary.EachRun[3].Command).To(Equal(1))
			Expect(summary.EachRun[3].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
			Expect(summary.EachRun[3].Failed).To(BeTrue())
			Expect(summary.EachRun[3].StartTime.UnixNano()).To(BeNumerically(">", summary.EachRun[2].StartTime.UnixNano()))

			Expect(summary.EachRun[4].Command).To(Equal(1))
			Expect(summary.EachRun[4].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
			Expect(summary.EachRun[4].Failed).To(BeFalse())
			Expect(summary.EachRun[4].StartTime.UnixNano()).To(BeNumerically(">", summary.EachRun[3].StartTime.UnixNano()))
		})

		Context("running multiple commands", func() {
			var (
				command1RunCount int
				command2RunCount int
			)

			BeforeEach(func() {
				command1RunCount = 0
				command2RunCount = 0

				commands = []string{"command 1", "command 2"}
				commandFunc = func(_ *exec.Cmd) error {
					command1RunCount++
					return nil
				}

				cmdRunner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "command",
					Args: []string{"2"},
				}, func(cmd *exec.Cmd) error {
					command2RunCount++
					return nil
				})
			})

			It("summarizes the commands it ran", func() {
				summary, err := countRunner.Run(1, cancelChan)
				Expect(err).NotTo(HaveOccurred())
				Expect(summary.Commands).To(HaveLen(2))

				command1 := summary.Commands[1]
				Expect(command1.Exec).To(Equal("command 1"))

				command2 := summary.Commands[2]
				Expect(command2.Exec).To(Equal("command 2"))
			})

			It("will eventually execute both", func() {
				_, err := countRunner.Run(1, cancelChan)
				Expect(err).NotTo(HaveOccurred())

				Expect(command2RunCount).To(BeNumerically(">", 0))
				Expect(command1RunCount).To(BeNumerically(">", 0))
			})
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
				}, 3*time.Second).Should(BeTrue())

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
				Expect(summary.EachRun[0].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
				Expect(summary.EachRun[0].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))

				Expect(summary.EachRun[1].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
				Expect(summary.EachRun[1].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))

				Expect(summary.EachRun[2].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
				Expect(summary.EachRun[2].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))

				Expect(summary.EachRun[3].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
				Expect(summary.EachRun[3].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))

				Expect(summary.EachRun[4].Duration).To(BeNumerically("~", 10*time.Millisecond, 5*time.Millisecond))
				Expect(summary.EachRun[4].StartTime.UnixNano()).To(BeNumerically("~", start.UnixNano(), time.Millisecond))
			})
		})
	})
})
