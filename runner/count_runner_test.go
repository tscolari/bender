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
		count       int
		cmdRunner   *fake_command_runner.FakeCommandRunner
		countRunner *runner.CountRunner

		commandFunc func(cmd *exec.Cmd) error
	)

	BeforeEach(func() {
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
		countRunner = runner.NewCountRunner(cmdRunner, count, args...)
	})

	Describe("Run", func() {
		It("calls the command runner with correct arguments", func() {
			_, err := countRunner.Run()
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

			summary, err := countRunner.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(summary.Duration).To(BeNumerically("~", time.Duration(count)*(10*time.Millisecond), 2*time.Millisecond))
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

			summary, err := countRunner.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(summary.SuccessCounter).To(Equal(count / 2))
			Expect(summary.ErrorCounter).To(Equal(count / 2))
		})
	})
})
