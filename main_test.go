package main_test

import (
	"encoding/json"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tscolari/bender/runner"
)

var _ = Describe("Main", func() {
	It("returns the correct summary", func() {
		outputBuffer := gbytes.NewBuffer()
		cmd := exec.Command(BenderBinPath, "--count", "3", "sleep", "1")
		session, err := gexec.Start(cmd, outputBuffer, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 4*time.Second).Should(gexec.Exit(0))

		var summary runner.Summary
		Expect(json.NewDecoder(outputBuffer).Decode(&summary)).ToNot(HaveOccurred())

		Expect(summary.Duration).To(BeNumerically("~", 3*time.Second, 10*time.Millisecond))
		Expect(summary.SuccessCounter).To(Equal(3))
		Expect(summary.ErrorCounter).To(BeZero())
	})
})
