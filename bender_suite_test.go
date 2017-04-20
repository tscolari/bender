package main_test

import (
	"encoding/json"
	"errors"
	"os/exec"
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

func RunBender(args ...string) (runner.Summary, error) {
	session, err := RunBenderSession(args...)
	if err != nil {
		return runner.Summary{}, err
	}
	Eventually(session, 5*time.Second).Should(gexec.Exit())

	if session.ExitCode() != 0 {
		return runner.Summary{}, errors.New(string(session.Err.Contents()))
	}

	return OutputToSummary(session.Out.Contents()), nil
}

func RunBenderSession(args ...string) (*gexec.Session, error) {
	outputBuffer := gbytes.NewBuffer()
	errorBuffer := gbytes.NewBuffer()

	cmd := exec.Command(BenderBinPath, args...)
	session, err := gexec.Start(cmd, outputBuffer, errorBuffer)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func OutputToSummary(output []byte) runner.Summary {
	var summary runner.Summary
	Expect(json.Unmarshal(output, &summary)).To(Succeed())
	return summary
}
