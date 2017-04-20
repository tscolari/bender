package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	It("returns the correct summary", func() {
		summary, err := RunBender("--count", "3", "--command", "sleep 1")
		Expect(err).NotTo(HaveOccurred())

		Expect(summary.Duration).To(BeNumerically("~", 3*time.Second, 100*time.Millisecond))
		Expect(summary.SuccessCounter).To(Equal(3))
		Expect(summary.ErrorCounter).To(BeZero())
	})

	It("exits gracefully when killed", func() {
		sess, err := RunBenderSession("--count", "500", "--command", "sleep 0.1")
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(time.Second)
		sess.Terminate()
		Eventually(sess).Should(gexec.Exit(0))

		summary := OutputToSummary(sess.Out.Contents())
		Expect(summary.SuccessCounter).To(BeNumerically(">", 1))
	})

	Context("when no type of runner is specified", func() {
		It("returns an error", func() {
			sess, err := RunBenderSession("--command", "sleep 0.1")
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(1))
			Eventually(sess.Err).Should(gbytes.Say("no runner detected"))
		})
	})

	Context("multiple commands", func() {
		It("randomizes between all the given commands", func() {
			summary, err := RunBender("--count", "5", "--command", "sleep 0", "--command", "sleep 1")
			Expect(err).NotTo(HaveOccurred())
			Expect(summary.Commands).To(HaveLen(2))

			Expect(summary.Commands[1].Exec).To(Equal("sleep 0"))
			Expect(summary.Commands[2].Exec).To(Equal("sleep 1"))

			Expect(summary.Duration).To(BeNumerically("~", time.Duration(summary.Commands[2].RunCount)*time.Second, 500*time.Millisecond))
			Expect(summary.SuccessCounter).To(Equal(5))
			Expect(summary.ErrorCounter).To(BeZero())
		})
	})

	Context("when the command fails", func() {
		It("returns the correct summary with the errors counter", func() {
			summary, err := RunBender("--count", "3", "--command", "do-not-exist")
			Expect(err).NotTo(HaveOccurred())

			Expect(summary.SuccessCounter).To(BeZero())
			Expect(summary.ErrorCounter).To(Equal(3))
		})
	})

	Context("when `--interval` is also provided", func() {
		It("returns an error", func() {
			_, err := RunBender("--count", "3", "--interval", "1s", "--command", "sleep 1")
			Expect(err).To(MatchError("can't use `--count` and `--interval` at the same time"))
		})
	})

	Context("Parallel runs", func() {
		It("returns the correct sumary", func() {
			summary, err := RunBender("--count", "3", "--concurrency", "3", "--command", "sleep 1")
			Expect(err).NotTo(HaveOccurred())

			Expect(summary.Duration).To(BeNumerically("~", 1*time.Second, 10*time.Millisecond))
			Expect(summary.SuccessCounter).To(Equal(3))
			Expect(summary.ErrorCounter).To(BeZero())
		})
	})

	Context("when --keep-running is provided", func() {
		It("runs until it gets cancelled", func() {
			sess, err := RunBenderSession("--keep-running", "--command", "sleep 0.1")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)
			Consistently(sess, 1*time.Second).ShouldNot(gexec.Exit())

			sess.Terminate()
			Eventually(sess).Should(gexec.Exit(0))

			summary := OutputToSummary(sess.Out.Contents())
			Expect(summary.Duration).To(BeNumerically("~", 2*time.Second, 500*time.Millisecond))
			Expect(summary.SuccessCounter).To(BeNumerically("~", 20, 2))
		})

		Context("and --interval is provided", func() {
			It("runs the commands in the given interval until it gets cancelled", func() {
				sess, err := RunBenderSession("--keep-running", "--interval", "0.5s", "--command", "sleep 0.1")
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(1 * time.Second)
				sess.Terminate()
				Eventually(sess).Should(gexec.Exit(0))

				summary := OutputToSummary(sess.Out.Contents())
				Expect(summary.Duration).To(BeNumerically("~", 1*time.Second, 500*time.Millisecond))
				Expect(summary.SuccessCounter).To(BeNumerically("~", 2, 1))
			})
		})

		Context("and --count is also provided", func() {
			It("fails to run", func() {
				_, err := RunBender("--count", "3", "--keep-running", "--command", "sleep 1")
				Expect(err).To(MatchError("can't use `--keep-running` and `--count` at the same time"))
			})
		})
	})
})
