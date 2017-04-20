package runner_test

import (
	"os/exec"
	"strings"
	"time"

	"code.cloudfoundry.org/commandrunner/fake_command_runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tscolari/bender/runner"
)

var _ = Describe("LoopRunner", func() {
	var (
		commands   []string
		cancelChan chan bool
		cmdRunner  *fake_command_runner.FakeCommandRunner
		loopRunner *runner.LoopRunner
		interval   time.Duration

		commandFunc func(cmd *exec.Cmd) error
	)

	BeforeEach(func() {
		interval = 0
		cancelChan = make(chan bool)
		commandFunc = nil
		cmdRunner = fake_command_runner.New()
		commands = []string{"hello world"}
	})

	JustBeforeEach(func() {
		loopRunner = runner.NewLoopRunnerWithCmdRunner(cmdRunner, interval, commands...)

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
		BeforeEach(func() {
			commandFunc = func(_ *exec.Cmd) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			}

		})

		It("runs until cancelled", func() {
			finished := make(chan bool, 1)
			go func() {
				defer GinkgoRecover()
				var err error
				_, err = loopRunner.Run(1, cancelChan)
				Expect(err).NotTo(HaveOccurred())
				close(finished)
			}()

			Consistently(finished, 100*time.Millisecond).ShouldNot(BeClosed())
			close(cancelChan)
			Eventually(finished).Should(BeClosed())
		})

		It("runs the correct commands", func() {
			finished := make(chan bool, 1)
			go func() {
				defer GinkgoRecover()
				var err error
				_, err = loopRunner.Run(1, cancelChan)
				Expect(err).NotTo(HaveOccurred())
				close(finished)
			}()

			Consistently(finished, 100*time.Millisecond).ShouldNot(BeClosed())
			close(cancelChan)

			Expect(len(cmdRunner.ExecutedCommands())).To(BeNumerically(">=", 10))
			for _, cmd := range cmdRunner.ExecutedCommands() {
				Expect(cmd.Args).To(Equal(strings.Split(commands[0], " ")))
			}
		})

		It("summarizes the total duration", func() {
			var summary runner.Summary
			finished := make(chan bool, 1)
			go func() {
				defer GinkgoRecover()
				var err error
				summary, err = loopRunner.Run(1, cancelChan)
				Expect(err).NotTo(HaveOccurred())
				close(finished)
			}()

			Consistently(finished, 100*time.Millisecond).ShouldNot(BeClosed())
			close(cancelChan)
			<-finished

			Expect(summary.Duration).To(BeNumerically("~", 100*time.Millisecond, 10*time.Millisecond))
		})

		It("summarizes the commands it ran", func() {
			var summary runner.Summary
			finished := make(chan bool, 1)
			go func() {
				defer GinkgoRecover()
				var err error
				summary, err = loopRunner.Run(1, cancelChan)
				Expect(err).NotTo(HaveOccurred())
				close(finished)
			}()

			Consistently(finished, 100*time.Millisecond).ShouldNot(BeClosed())
			close(cancelChan)
			<-finished

			Expect(summary.Commands).To(HaveLen(1))
			Expect(summary.Commands[1].Exec).To(Equal("hello world"))
		})

		Context("interval", func() {
			BeforeEach(func() {
				interval = 100 * time.Millisecond
				commandFunc = func(_ *exec.Cmd) error {
					time.Sleep(10 * time.Millisecond)
					return nil
				}
			})

			It("runs the command in an interval", func() {
				var summary runner.Summary
				finished := make(chan bool, 1)
				go func() {
					defer GinkgoRecover()
					var err error
					summary, err = loopRunner.Run(1, cancelChan)
					Expect(err).NotTo(HaveOccurred())
					close(finished)
				}()

				time.Sleep(1 * time.Second)
				close(cancelChan)
				<-finished

				Expect(summary.SuccessCounter).To(BeNumerically("~", 10, 2))
			})
		})
	})
})
