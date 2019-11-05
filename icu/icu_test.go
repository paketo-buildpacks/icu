package icu

import (
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	. "github.com/onsi/gomega"

	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitICU(t *testing.T) {
	spec.Run(t, "ICU", testICU, spec.Report(report.Terminal{}))
}

func testICU(t *testing.T, when spec.G, it spec.S) {
	var (
		factory        *test.BuildFactory
		stubICUFixture = filepath.Join("testdata", "stub-icu-dependency.tgz")
	)

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
		factory.AddDependencyWithVersion(Dependency, "65.1", stubICUFixture)
	})

	when("runtime.NewContributor", func() {
		it("returns true if a build plan exists and matching version is found", func() {
			factory.AddPlan(buildpackplan.Plan{Name: Dependency, Version: "65.1"})

			_, willContribute, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan exists and no matching version is found", func() {
			factory.AddPlan(buildpackplan.Plan{Name: Dependency, Version: "60.0"})

			_, willContribute, err := NewContributor(factory.Build)
			Expect(err).To(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})

		it("returns false if a build plan does not exist", func() {
			contributor, willContribute, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
			Expect(contributor).To(Equal(Contributor{}))
		})
	})

	when("Contribute", func() {
		it("installs the icu dependency and moves it on disk", func() {
			factory.AddPlan(buildpackplan.Plan{Name: Dependency, Version: "65.1"})

			ICUContributor, _, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(ICUContributor.Contribute()).To(Succeed())

			layer := factory.Build.Layers.Layer(Dependency)
			Expect(filepath.Join(layer.Root, "bin")).To(BeADirectory())
			Expect(filepath.Join(layer.Root, "include")).To(BeADirectory())
			Expect(filepath.Join(layer.Root, "lib")).To(BeADirectory())
		})

		it("contributes dotnet runtime to the build layer when included in the build plan", func() {
			factory.AddPlan(buildpackplan.Plan{
				Name:    Dependency,
				Version: "65.1",
				Metadata: buildpackplan.Metadata{
					"build": true,
				},
			})

			dotnetRuntimeContributor, _, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(dotnetRuntimeContributor.Contribute()).To(Succeed())

			layer := factory.Build.Layers.Layer(Dependency)
			Expect(layer).To(test.HaveLayerMetadata(true, false, false))
		})

		it("contributes dotnet runtime to the launch layer when included in the build plan", func() {
			factory.AddPlan(buildpackplan.Plan{
				Name:    Dependency,
				Version: "65.1",
				Metadata: buildpackplan.Metadata{
					"launch": true,
				},
			})

			dotnetRuntimeContributor, _, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(dotnetRuntimeContributor.Contribute()).To(Succeed())

			layer := factory.Build.Layers.Layer(Dependency)
			Expect(layer).To(test.HaveLayerMetadata(false, false, true))
		})
	})
}
