package main

import (
	"os"

	"github.com/paketo-buildpacks/icu"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	packit.Run(
		icu.Detect(),
		icu.Build(
			draft.NewPlanner(),
			postal.NewService(cargo.NewTransport()),
			icu.NewICULayerArranger(),
			chronos.DefaultClock,
			scribe.NewEmitter(os.Stdout),
		),
	)
}
