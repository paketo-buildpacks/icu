package icu

import "github.com/paketo-buildpacks/packit"

type PlanEntryResolver struct{}

func NewPlanEntryResolver() PlanEntryResolver {
	return PlanEntryResolver{}
}

func (p PlanEntryResolver) Resolve(entries []packit.BuildpackPlanEntry) packit.BuildpackPlanEntry {
	mergedEntry := packit.BuildpackPlanEntry{
		Name:     "icu",
		Metadata: map[string]interface{}{},
	}

	for _, e := range entries {
		for _, phase := range []string{"build", "launch"} {
			if e.Metadata[phase] == true {
				mergedEntry.Metadata[phase] = true
			}
		}
	}

	return mergedEntry
}
