package main

import (
	"os"

	"github.com/paketo-buildpacks/icu"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	entryResolver := icu.NewPlanEntryResolver()
	dependencyManager := postal.NewService(cargo.NewTransport())
	layerArranger := icu.NewICULayerArranger()
	logger := icu.NewLogEmitter(os.Stdout)
	packit.Run(
		icu.Detect(),
		icu.Build(
			entryResolver,
			dependencyManager,
			layerArranger,
			chronos.DefaultClock,
			logger,
		),
	)
}
