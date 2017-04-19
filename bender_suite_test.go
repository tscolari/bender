package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

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
