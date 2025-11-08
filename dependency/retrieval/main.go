package main

import (
	"github.com/paketo-buildpacks/icu/dependency/retrieval/components"
	"github.com/paketo-buildpacks/libdependency/retrieve"
)

func main() {
	fetcher := components.NewFetcher()
	generator := components.NewGenerator()
	retrieve.NewMetadata("icu", fetcher.GetIcuVersions, generator.GenerateMetadata)
}
