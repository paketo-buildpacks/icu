package main

import (
	"os"

	"github.com/paketo-buildpacks/icu"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	packit.Run(
		icu.Detect(),
		icu.Build(
			postal.NewService(cargo.NewTransport()),
			icu.NewICULayerArranger(),
			chronos.DefaultClock,
			scribe.NewEmitter(os.Stdout),
		),
	)
}
