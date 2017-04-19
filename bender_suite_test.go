package main_test

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tscolari/bender/runner"

	"testing"
)

var BenderBinPath string

func TestBender(t *testing.T) {

	SynchronizedBeforeSuite(func() []byte {
		binPath, err := gexec.Build("github.com/tscolari/bender")
		Expect(err).NotTo(HaveOccurred())

		return []byte(binPath)
	}, func(binPath []byte) {
		BenderBinPath = string(binPath)
	})

	SynchronizedAfterSuite(func() {}, func() {
		gexec.CleanupBuildArtifacts()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Bender Suite")
}

func RunBender(count int, parallel int, args ...string) runner.Summary {
	outputBuffer := gbytes.NewBuffer()

	args = append([]string{"--count", strconv.Itoa(count), "--concurrency", strconv.Itoa(parallel)}, args...)
	cmd := exec.Command(BenderBinPath, args...)
	session, err := gexec.Start(cmd, outputBuffer, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 5*time.Second).Should(gexec.Exit(0))

	var summary runner.Summary
	Expect(json.NewDecoder(outputBuffer).Decode(&summary)).ToNot(HaveOccurred())

	return summary
}
