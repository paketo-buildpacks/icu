package icu_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	icu "github.com/paketo-buildpacks/icu"
	"github.com/paketo-buildpacks/icu/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
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
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
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

		layerArranger = &fakes.LayerArranger{}

		buffer = bytes.NewBuffer(nil)
		logEmitter := icu.NewLogEmitter(buffer)

		timestamp = time.Now()
		clock := chronos.NewClock(func() time.Time {
			return timestamp
		})

		build = icu.Build(entryResolver, dependencyManager, layerArranger, clock, logEmitter)
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
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{Name: "icu"},
				},
			},
			Layers: []packit.Layer{
				{
					Name:      "icu",
					Path:      filepath.Join(layersDir, "icu"),
					SharedEnv: packit.Environment{},
					BuildEnv:  packit.Environment{},
					LaunchEnv: packit.Environment{},
					Build:     false,
					Launch:    false,
					Cache:     false,
					Metadata: map[string]interface{}{
						icu.DependencyCacheKey: "icu-dependency-sha",
						"built_at":             timestamp.Format(time.RFC3339Nano),
					},
				},
			},
		}))

		Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
			{
				Name: "icu",
			},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("icu"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("*"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.InstallCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "icu",
			Name:    "icu-dependency-name",
			SHA256:  "icu-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "icu-dependency-uri",
			Version: "icu-dependency-version",
		}))
		Expect(dependencyManager.InstallCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.InstallCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "icu")))

		Expect(layerArranger.ArrangeCall.Receives.Path).To(Equal(filepath.Join(layersDir, "icu")))
	})

	context("when the plan entry requires the dependency during the build and launch phases", func() {
		it.Before(func() {
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "icu",
				Metadata: map[string]interface{}{
					"build":  true,
					"launch": true,
					"cache":  true,
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
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{Name: "icu"},
					},
				},
				Layers: []packit.Layer{
					{
						Name:      "icu",
						Path:      filepath.Join(layersDir, "icu"),
						SharedEnv: packit.Environment{},
						BuildEnv:  packit.Environment{},
						LaunchEnv: packit.Environment{},
						Build:     true,
						Launch:    true,
						Cache:     true,
						Metadata: map[string]interface{}{
							icu.DependencyCacheKey: "icu-dependency-sha",
							"built_at":             timestamp.Format(time.RFC3339Nano),
						},
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("when the ICU layer cannot be retrieved", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(layersDir, "icu.toml"), nil, 0000)
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
			dependencyManager.InstallCall.Returns.Error = errors.New("failed to install dependency")
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
