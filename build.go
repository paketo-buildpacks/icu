package icu

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface LayerArranger --output fakes/layer_arranger.go
type LayerArranger interface {
	Arrange(path string) error
}

//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go
type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

func Build(dependencyManager DependencyManager,
	layerArranger LayerArranger,
	sbomGenerator SBOMGenerator,
	clock chronos.Clock,
	logger scribe.Emitter,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		layer, err := context.Layers.Get(ICULayerName)
		if err != nil {
			return packit.BuildResult{}, err
		}

		planner := draft.NewPlanner()
		entry, _ := planner.Resolve("icu", context.Plan.Entries, nil)

		dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, "*", context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bom := dependencyManager.GenerateBillOfMaterials(dependency)
		launch, build := planner.MergeLayerTypes("icu", context.Plan.Entries)

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = bom
		}

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = bom
		}

		cachedSHA, ok := layer.Metadata[DependencyCacheKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", layer.Path)
			logger.Break()

			layer.Launch, layer.Build, layer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{layer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		logger.Process("Executing build process")

		layer, err = layer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		layer.Launch, layer.Build, layer.Cache = launch, build, build

		logger.Subprocess("Installing ICU")

		duration, err := clock.Measure(func() error {
			return dependencyManager.Deliver(dependency, context.CNBPath, layer.Path, context.Platform.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		// LayerArranger is a stop gap until we can get the dependency artifact
		// restructured to remove the top two directories
		err = layerArranger.Arrange(layer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.GeneratingSBOM(layer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, layer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		layer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		layer.Metadata = map[string]interface{}{
			DependencyCacheKey: dependency.SHA256,
		}

		return packit.BuildResult{
			Layers: []packit.Layer{layer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
