package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	It("returns the correct summary", func() {
		summary := RunBender(3, 1, "--command", "sleep 1")

		Expect(summary.Duration).To(BeNumerically("~", 3*time.Second, 100*time.Millisecond))
		Expect(summary.SuccessCounter).To(Equal(3))
		Expect(summary.ErrorCounter).To(BeZero())
	})

	Context("multiple commands", func() {
		It("randomizes between all the given commands", func() {
			summary := RunBender(5, 1, "--command", "sleep 0", "--command", "sleep 1")
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
			summary := RunBender(3, 1, "--command", "do-not-exist")

			Expect(summary.SuccessCounter).To(BeZero())
			Expect(summary.ErrorCounter).To(Equal(3))
		})
	})

	Context("Parallel runs", func() {
		It("returns the correct sumary", func() {
			summary := RunBender(3, 3, "--command", "sleep 1")

			Expect(summary.Duration).To(BeNumerically("~", 1*time.Second, 10*time.Millisecond))
			Expect(summary.SuccessCounter).To(Equal(3))
			Expect(summary.ErrorCounter).To(BeZero())
		})
	})
})
