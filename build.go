package icu

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve([]packit.BuildpackPlanEntry) packit.BuildpackPlanEntry
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//go:generate faux --interface LayerArranger --output fakes/layer_arranger.go
type LayerArranger interface {
	Arrange(path string) error
}

func Build(entryResolver EntryResolver,
	dependencyManager DependencyManager,
	layerArranger LayerArranger,
	clock chronos.Clock,
	logger LogEmitter,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		icuLayer, err := context.Layers.Get("icu")
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Executing build process")

		err = icuLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		icuEntry := entryResolver.Resolve(context.Plan.Entries)

		icuLayer.Build = icuEntry.Metadata["build"] == true
		icuLayer.Cache = icuEntry.Metadata["build"] == true
		icuLayer.Launch = icuEntry.Metadata["launch"] == true

		dep, err := dependencyManager.Resolve(
			filepath.Join(context.CNBPath, "buildpack.toml"),
			icuEntry.Name,
			"*",
			context.Stack,
		)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Subprocess("Installing ICU")

		duration, err := clock.Measure(func() error {
			return dependencyManager.Install(dep, context.CNBPath, icuLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		// LayerArranger is a stop gap until we can get the dependency artifact
		// restructured to remove the top two directories
		err = layerArranger.Arrange(icuLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Plan:   context.Plan,
			Layers: []packit.Layer{icuLayer},
		}, nil
	}
}
