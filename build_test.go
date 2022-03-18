package icu_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	icu "github.com/paketo-buildpacks/icu"
	"github.com/paketo-buildpacks/icu/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"

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

		entryResolver     *fakes.EntryResolver
		dependencyManager *fakes.DependencyManager
		layerArranger     *fakes.LayerArranger

		buffer *bytes.Buffer

		timestamp time.Time

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "icu",
		}

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:      "icu",
			Name:    "icu-dependency-name",
			SHA256:  "icu-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "icu-dependency-uri",
			Version: "icu-dependency-version",
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

		timestamp = time.Now()
		clock := chronos.NewClock(func() time.Time {
			return timestamp
		})

		build = icu.Build(entryResolver, dependencyManager, layerArranger, clock, scribe.NewEmitter(buffer))
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that includes ICU", func() {
		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Stack:      "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
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
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Layers: []packit.Layer{
				{
					Name:             "icu",
					Path:             filepath.Join(layersDir, "icu"),
					SharedEnv:        packit.Environment{},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					ProcessLaunchEnv: map[string]packit.Environment{},
					Build:            false,
					Launch:           false,
					Cache:            false,
					Metadata: map[string]interface{}{
						icu.DependencyCacheKey: "icu-dependency-sha",
						"built_at":             timestamp.Format(time.RFC3339Nano),
					},
				},
			},
		}))

		Expect(entryResolver.ResolveCall.Receives.Name).To(Equal("icu"))
		Expect(entryResolver.ResolveCall.Receives.Entries).To(Equal([]packit.BuildpackPlanEntry{
			{Name: "icu"},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("icu"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("*"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
			{
				ID:      "icu",
				Name:    "icu-dependency-name",
				SHA256:  "icu-dependency-sha",
				Stacks:  []string{"some-stack"},
				URI:     "icu-dependency-uri",
				Version: "icu-dependency-version",
			},
		}))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "icu",
			Name:    "icu-dependency-name",
			SHA256:  "icu-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "icu-dependency-uri",
			Version: "icu-dependency-version",
		}))
		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "icu")))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("platform"))

		Expect(layerArranger.ArrangeCall.Receives.Path).To(Equal(filepath.Join(layersDir, "icu")))
	})

	context("when the plan entry requires the dependency during the build and launch phases", func() {
		it.Before(func() {
			entryResolver.MergeLayerTypesCall.Returns.Launch = true
			entryResolver.MergeLayerTypesCall.Returns.Build = true
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "icu",
				Metadata: map[string]interface{}{
					"build":  true,
					"launch": true,
				},
			}
		})

		it("makes a layer available in those phases", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "icu",
							Metadata: map[string]interface{}{
								"build":  true,
								"launch": true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name:             "icu",
						Path:             filepath.Join(layersDir, "icu"),
						SharedEnv:        packit.Environment{},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            true,
						Launch:           true,
						Cache:            true,
						Metadata: map[string]interface{}{
							icu.DependencyCacheKey: "icu-dependency-sha",
							"built_at":             timestamp.Format(time.RFC3339Nano),
						},
					},
				},
				Build: packit.BuildMetadata{
					BOM: []packit.BOMEntry{
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
					},
				},
				Launch: packit.LaunchMetadata{
					BOM: []packit.BOMEntry{
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
					},
				},
			}))
		})
	})

	context("when there is a cache match in the layer metadata", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(layersDir, "icu.toml"),
				[]byte("[metadata]\ndependency-sha = \"icu-dependency-sha\"\n"), 0600)
			Expect(err).NotTo(HaveOccurred())

			entryResolver.MergeLayerTypesCall.Returns.Launch = false
			entryResolver.MergeLayerTypesCall.Returns.Build = true
		})

		it("reuses the layer", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{Name: "icu"},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Layers: []packit.Layer{
					{
						Name:             "icu",
						Path:             filepath.Join(layersDir, "icu"),
						SharedEnv:        packit.Environment{},
						BuildEnv:         packit.Environment{},
						LaunchEnv:        packit.Environment{},
						ProcessLaunchEnv: map[string]packit.Environment{},
						Build:            true,
						Launch:           false,
						Cache:            true,
						Metadata: map[string]interface{}{
							icu.DependencyCacheKey: "icu-dependency-sha",
						},
					},
				},
				Build: packit.BuildMetadata{
					BOM: []packit.BOMEntry{
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
					},
				},
			}))

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
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "icu",
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
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
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "icu",
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("could not remove file")))
			})

		})

		context("when the dependencyManager Resolve fails", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("fails with the error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name: "icu",
							},
						},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})
	})

	context("when the dependencyManager Install fails", func() {
		it.Before(func() {
			dependencyManager.DeliverCall.Returns.Error = errors.New("failed to install dependency")
		})

		it("fails with the error", func() {
			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "icu",
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).To(MatchError("failed to install dependency"))
		})
	})

	context("when the layerArranger Arrange fails", func() {
		it.Before(func() {
			layerArranger.ArrangeCall.Returns.Error = errors.New("failed to arrange layer")
		})

		it("fails with the error", func() {
			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "icu",
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).To(MatchError("failed to arrange layer"))
		})
	})
}
