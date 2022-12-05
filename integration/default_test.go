package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		pack       occam.Pack
		docker     occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building a simple app", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
			source    string
			sbomDir   string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			sbomDir, err = os.MkdirTemp("", "sbom")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chmod(sbomDir, os.ModePerm)).To(Succeed())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
			Expect(os.RemoveAll(sbomDir)).To(Succeed())
		})

		it("builds an oci image with icu installed", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					buildpack,
					buildPlanBuildpack,
				).
				WithSBOMOutputDir(sbomDir).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs.String()).To(ContainSubstring(buildpackInfo.Buildpack.Name))
			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving ICU version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected ICU version \(using <unknown>\): \d+\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing ICU`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))

			container, err = docker.Container.Run.
				WithCommand(fmt.Sprintf("icuinfo && test -f /layers/%s/icu/lib/libicudata.so && echo 'libicudata.so exists' && sleep infinity",
					strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(
				And(
					ContainSubstring("International Components for Unicode for C/C++"),
					ContainSubstring("libicudata.so exists"),
				),
			)

			contents, err := os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", "sbom.legacy.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring(`"name":"ICU"`))

			// check that all required SBOM files are present
			Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "icu", "sbom.cdx.json")).To(BeARegularFile())
			Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "icu", "sbom.spdx.json")).To(BeARegularFile())
			Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "icu", "sbom.syft.json")).To(BeARegularFile())

			// check an SBOM file
			contents, err = os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "icu", "sbom.cdx.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring(`"name": "ICU"`))
		})

		context("a version is requested via the dotnet-31 version source", func() {
			it("builds an oci image with requested icu version", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "required_version_app"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						buildpack,
						buildPlanBuildpack,
					).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs.String()).To(ContainSubstring(buildpackInfo.Buildpack.Name))
				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
					"  Resolving ICU version",
					"    Candidate version sources (in priority order):",
					"      dotnet-31 -> \"70.*\"",
					"",
					MatchRegexp(`    Selected ICU version \(using dotnet-31\): \d+\.\d+`),
					"",
					"  Executing build process",
					MatchRegexp(`    Installing ICU`),
					MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				))

				container, err = docker.Container.Run.
					WithCommand("icuinfo").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(
					And(
						ContainSubstring("International Components for Unicode for C/C++"),
						ContainSubstring(`<param name="version">70.1</param>`),
					),
				)
			})
		})
	})
}
