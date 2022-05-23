package acceptance_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/paketo-buildpacks/packit/pexec"
)

const lifecycleVersion = "0.14.0"

func testBuildpackIntegration(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		buildPlanBuildpack string
		goDistBuildpack    string

		builderConfigFilepath string

		pack   occam.Pack
		docker occam.Docker
		source string
		name   string

		image     occam.Image
		container occam.Container
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()

		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		buildpackStore := occam.NewBuildpackStore()

		buildPlanBuildpack, err = buildpackStore.Get.
			Execute("github.com/paketo-community/build-plan")
		Expect(err).NotTo(HaveOccurred())

		goDistBuildpack, err = buildpackStore.Get.
			WithVersion("1.2.3").
			Execute("$HOME/workspace/paketo-buildpacks/go-dist")
		Expect(err).NotTo(HaveOccurred())

		source, err = occam.Source(filepath.Join("integration", "testdata", "simple_app"))
		Expect(err).NotTo(HaveOccurred())

		builderConfigFile, err := os.CreateTemp("", "builder.toml")
		Expect(err).NotTo(HaveOccurred())
		builderConfigFilepath = builderConfigFile.Name()

		_, err = fmt.Fprintf(builderConfigFile, `
[lifecycle]
  version = "%s"

[stack]
  build-image = "%s:latest"
  id = "io.buildpacks.stacks.jammy.tiny"
  run-image = "%s:latest"
`,
			lifecycleVersion,
			stack.BuildImageID,
			stack.RunImageID,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(docker.Pull.Execute(fmt.Sprintf("buildpacksio/lifecycle:%s", lifecycleVersion))).To(Succeed())

		// Create builder from stacks

		skopeo := pexec.NewExecutable("skopeo")

		err = skopeo.Execute(pexec.Execution{
			Args: []string{
				"copy",
				fmt.Sprintf("oci-archive://%s", stack.BuildArchive),
				fmt.Sprintf("docker-daemon:%s:latest", stack.BuildImageID),
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = skopeo.Execute(pexec.Execution{
			Args: []string{
				"copy",
				fmt.Sprintf("oci-archive://%s", stack.RunArchive),
				fmt.Sprintf("docker-daemon:%s:latest", stack.RunImageID),
			},
		})
		Expect(err).NotTo(HaveOccurred())

		pPackBuffer := bytes.NewBuffer(nil)

		pPack := pexec.NewExecutable("pack")
		err = pPack.Execute(pexec.Execution{
			Stdout: pPackBuffer,
			Stderr: pPackBuffer,
			Args: []string{
				"builder",
				"create",
				"my-builder",
				fmt.Sprintf("--config=%s", builderConfigFilepath),
			},
		})
		Expect(err).NotTo(HaveOccurred(), pPackBuffer.String())
	})

	it.After(func() {
		// Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
		// Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
		// Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(docker.Image.Remove.Execute("my-builder")).To(Succeed())
		Expect(docker.Image.Remove.Execute(stack.BuildImageID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(stack.RunImageID)).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
		Expect(os.RemoveAll(builderConfigFilepath)).To(Succeed())

		// Clean up builder
	})

	it("builds an app with a buildpack", func() {

		var err error
		var logs fmt.Stringer
		image, logs, err = pack.WithNoColor().Build.
			WithPullPolicy("never").
			WithBuildpacks(
				goDistBuildpack,
				buildPlanBuildpack,
			).
			WithEnv(map[string]string{
				"BP_LOG_LEVEL": "DEBUG",
			}).
			WithBuilder("my-builder").
			Execute(name, source)
		Expect(err).ToNot(HaveOccurred(), logs.String)

		// Ensure go is installed correctly
		container, err = docker.Container.Run.
			WithDirect().
			WithCommand("go").
			WithCommandArgs([]string{"run", "main.go"}).
			WithEnv(map[string]string{"PORT": "8080"}).
			WithPublish("8080").
			WithPublishAll().
			Execute(image.ID)
		Expect(err).NotTo(HaveOccurred())

		Eventually(container).Should(BeAvailable())
		Eventually(container).Should(Serve(ContainSubstring("go1.17")).OnPort(8080))
	})
}
