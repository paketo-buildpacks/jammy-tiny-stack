package acceptance_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var stack struct {
	BuildArchive string
	RunArchive   string
	BuildImageID string
	RunImageID   string
}

// var skopeo pexec.Executable

func by(_ string, f func()) { f() }

func TestAcceptance(t *testing.T) {
	format.MaxLength = 0
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs(".")
	Expect(err).ToNot(HaveOccurred())

	stack.BuildArchive = filepath.Join(root, "build", "build.oci")
	// build, err := occam.RandomName()
	// Expect(err).NotTo(HaveOccurred())
	// stack.BuildImageID = build
	stack.BuildImageID = "stack-build-image-id"

	stack.RunArchive = filepath.Join(root, "build", "run.oci")
	// run, err := occam.RandomName()
	// Expect(err).NotTo(HaveOccurred())
	// stack.RunImageID = run
	stack.RunImageID = "stack-run-image-id"

	// err = skopeo.Execute(pexec.Execution{
	// 	Args: []string{
	// 		"copy",
	// 		fmt.Sprintf("oci-archive://%s", stack.RunArchive),
	// 		fmt.Sprintf("docker-daemon:%s:latest", stack.RunImageID),
	// 	},
	// })
	// Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Acceptance", spec.Report(report.Terminal{}), spec.Parallel())
	// suite("Metadata", testMetadata)
	suite("BuildpackIntegration", testBuildpackIntegration)

	suite.Run(t)
}
