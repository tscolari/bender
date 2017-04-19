package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	It("returns the correct summary", func() {
		summary := RunBender(3, 1, "sleep", "1")

		Expect(summary.Duration).To(BeNumerically("~", 3*time.Second, 10*time.Millisecond))
		Expect(summary.SuccessCounter).To(Equal(3))
		Expect(summary.ErrorCounter).To(BeZero())
	})

	Context("when the command fails", func() {
		It("returns the correct summary with the errors counter", func() {
			summary := RunBender(3, 1, "do-not-exist")

			Expect(summary.SuccessCounter).To(BeZero())
			Expect(summary.ErrorCounter).To(Equal(3))
		})
	})

	Context("Parallel runs", func() {
		It("returns the correct sumary", func() {
			summary := RunBender(3, 3, "sleep", "1")

			Expect(summary.Duration).To(BeNumerically("~", 1*time.Second, 10*time.Millisecond))
			Expect(summary.SuccessCounter).To(Equal(3))
			Expect(summary.ErrorCounter).To(BeZero())
		})
	})
})
