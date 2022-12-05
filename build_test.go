package icu_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/icu"
	"github.com/paketo-buildpacks/icu/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"

	//nolint Ignore SA1019, informed usage of deprecated package
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string

		dependencyManager *fakes.DependencyManager
		layerArranger     *fakes.LayerArranger
		sbomGenerator     *fakes.SBOMGenerator

		buffer *bytes.Buffer

		buildContext packit.BuildContext
		build        packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:       "icu",
			Name:     "ICU",
			Checksum: "icu-dependency-sha",
			Stacks:   []string{"some-stack"},
			URI:      "icu-dependency-uri",
			Version:  "icu-dependency-version",
		}
		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "icu",
				Metadata: paketosbom.BOMMetadata{
					Version: "icu-dependency-version",
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "icu-dependency-sha",
					},
					URI: "icu-dependency-uri",
				},
			},
		}

		layerArranger = &fakes.LayerArranger{}

		buffer = bytes.NewBuffer(nil)

		buildContext = packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Stack:      "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "Some Buildpack",
				Version:     "some-version",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
			},
			Platform: packit.Platform{Path: "platform"},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "icu",
					},
				},
			},
			Layers: packit.Layers{Path: layersDir},
		}

		sbomGenerator = &fakes.SBOMGenerator{}
		sbomGenerator.GenerateFromDependencyCall.Returns.SBOM = sbom.SBOM{}

		build = icu.Build(
			dependencyManager,
			layerArranger,
			sbomGenerator,
			chronos.DefaultClock,
			scribe.NewEmitter(buffer))
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that includes ICU", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("icu"))
		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "icu")))
		Expect(layer.Metadata).To(Equal(map[string]interface{}{
			"dependency-checksum": "icu-dependency-sha",
		}))

		Expect(layer.Build).To(BeFalse())
		Expect(layer.Launch).To(BeFalse())
		Expect(layer.Cache).To(BeFalse())

		Expect(layer.SBOM.Formats()).To(Equal([]packit.SBOMFormat{
			{
				Extension: sbom.Format(sbom.CycloneDXFormat).Extension(),
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.CycloneDXFormat),
			},
			{
				Extension: sbom.Format(sbom.SPDXFormat).Extension(),
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.SPDXFormat),
			},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("icu"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("*"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
			{
				ID:       "icu",
				Name:     "ICU",
				Checksum: "icu-dependency-sha",
				Stacks:   []string{"some-stack"},
				URI:      "icu-dependency-uri",
				Version:  "icu-dependency-version",
			},
		}))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:       "icu",
			Name:     "ICU",
			Checksum: "icu-dependency-sha",
			Stacks:   []string{"some-stack"},
			URI:      "icu-dependency-uri",
			Version:  "icu-dependency-version",
		}))
		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "icu")))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("platform"))

		Expect(layerArranger.ArrangeCall.Receives.Path).To(Equal(filepath.Join(layersDir, "icu")))

		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:       "icu",
			Name:     "ICU",
			Checksum: "icu-dependency-sha",
			Stacks:   []string{"some-stack"},
			URI:      "icu-dependency-uri",
			Version:  "icu-dependency-version",
		}))
		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dir).To(Equal(filepath.Join(layersDir, "icu")))
	})

	context("when the plan entry requires multiple specific versions", func() {
		it.Before(func() {
			buildContext.Plan.Entries = []packit.BuildpackPlanEntry{
				{
					Name: "icu",
					Metadata: map[string]interface{}{
						"launch":         true,
						"version":        "70.*",
						"version-source": "dotnet-31",
					},
				},
				{
					Name: "icu",
					Metadata: map[string]interface{}{
						"launch":         true,
						"version":        "71.1.*",
						"version-source": "random-source",
					},
				},
			}
		})

		it("chooses highest priority (dotnet-31) version source", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("icu"))
			Expect(layer.Path).To(Equal(filepath.Join(layersDir, "icu")))
			Expect(layer.Metadata).To(Equal(map[string]interface{}{
				"dependency-checksum": "icu-dependency-sha",
			}))
			Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
			Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("icu"))
			Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("70.*"))
			Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))
		})
	})

	context("when the plan entry requires the dependency during the build and launch phases", func() {
		it.Before(func() {
			buildContext.Plan.Entries[0].Metadata = map[string]interface{}{
				"build":  true,
				"launch": true,
			}
		})

		it("makes a layer available in those phases", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("icu"))
			Expect(layer.Path).To(Equal(filepath.Join(layersDir, "icu")))
			Expect(layer.Metadata).To(Equal(map[string]interface{}{
				"dependency-checksum": "icu-dependency-sha",
			}))

			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeTrue())
			Expect(layer.Cache).To(BeTrue())

			Expect(result.Build.BOM).To(HaveLen(1))
			buildBOMEntry := result.Build.BOM[0]
			Expect(buildBOMEntry.Name).To(Equal("icu"))
			Expect(buildBOMEntry.Metadata).To(Equal(paketosbom.BOMMetadata{
				Version: "icu-dependency-version",
				Checksum: paketosbom.BOMChecksum{
					Algorithm: paketosbom.SHA256,
					Hash:      "icu-dependency-sha",
				},
				URI: "icu-dependency-uri",
			}))

			Expect(result.Launch.BOM).To(HaveLen(1))
			launchBOMEntry := result.Launch.BOM[0]
			Expect(launchBOMEntry.Name).To(Equal("icu"))
			Expect(launchBOMEntry.Metadata).To(Equal(paketosbom.BOMMetadata{
				Version: "icu-dependency-version",
				Checksum: paketosbom.BOMChecksum{
					Algorithm: paketosbom.SHA256,
					Hash:      "icu-dependency-sha",
				},
				URI: "icu-dependency-uri",
			}))
		})
	})

	context("when there is a cache match in the layer metadata", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(layersDir, "icu.toml"),
				[]byte("[metadata]\ndependency-checksum = \"icu-dependency-sha\"\n"), 0600)
			Expect(err).NotTo(HaveOccurred())

			buildContext.Plan.Entries[0].Metadata = map[string]interface{}{
				"launch": false,
				"build":  true,
			}
		})

		it("reuses the layer", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("icu"))
			Expect(layer.Path).To(Equal(filepath.Join(layersDir, "icu")))
			Expect(layer.Metadata).To(Equal(map[string]interface{}{
				"dependency-checksum": "icu-dependency-sha",
			}))

			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeFalse())
			Expect(layer.Cache).To(BeTrue())

			Expect(result.Build.BOM).To(HaveLen(1))
			buildBOMEntry := result.Build.BOM[0]
			Expect(buildBOMEntry.Name).To(Equal("icu"))
			Expect(buildBOMEntry.Metadata).To(Equal(paketosbom.BOMMetadata{
				Version: "icu-dependency-version",
				Checksum: paketosbom.BOMChecksum{
					Algorithm: paketosbom.SHA256,
					Hash:      "icu-dependency-sha",
				},
				URI: "icu-dependency-uri",
			}))

			Expect(result.Launch.BOM).To(HaveLen(0))

			Expect(dependencyManager.DeliverCall.CallCount).To(Equal(0))
		})
	})

	context("failure cases", func() {
		context("when the ICU layer cannot be retrieved", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(layersDir, "icu.toml"), nil, 0000)
				Expect(err).NotTo(HaveOccurred())
			})

			it("fails with the error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to parse layer content metadata")))
			})
		})

		context("when the ICU layer cannot be reset", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layersDir, "icu", "something"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(filepath.Join(layersDir, "icu"), 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(layersDir, "icu"), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("could not remove file")))
			})

		})

		context("when the dependencyManager Resolve fails", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("fails with the error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})
	})

	context("when the dependencyManager Install fails", func() {
		it.Before(func() {
			dependencyManager.DeliverCall.Returns.Error = errors.New("failed to install dependency")
		})

		it("fails with the error", func() {
			_, err := build(buildContext)
			Expect(err).To(MatchError("failed to install dependency"))
		})
	})

	context("when the layerArranger Arrange fails", func() {
		it.Before(func() {
			layerArranger.ArrangeCall.Returns.Error = errors.New("failed to arrange layer")
		})

		it("fails with the error", func() {
			_, err := build(buildContext)
			Expect(err).To(MatchError("failed to arrange layer"))
		})
	})

	context("when generating the SBOM returns an error", func() {
		it.Before(func() {
			sbomGenerator.GenerateFromDependencyCall.Returns.Error = errors.New("failed to generate SBOM")
		})

		it("returns an error", func() {
			_, err := build(buildContext)
			Expect(err).To(MatchError(ContainSubstring("failed to generate SBOM")))
		})
	})

	context("when formatting the SBOM returns an error", func() {
		it.Before(func() {
			buildContext.BuildpackInfo.SBOMFormats = []string{"random-format"}
		})

		it("returns an error", func() {
			_, err := build(buildContext)
			Expect(err).To(MatchError("unsupported SBOM format: 'random-format'"))
		})
	})
}
