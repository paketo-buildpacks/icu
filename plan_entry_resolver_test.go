package icu_test

import (
	"testing"

	icu "github.com/paketo-buildpacks/icu"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanEntryResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		resolver icu.PlanEntryResolver
	)

	it.Before(func() {
		resolver = icu.NewPlanEntryResolver()
	})

	context("when the entry flags differ", func() {
		it("ORs together the flags to produce the entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "icu",
					Metadata: map[string]interface{}{
						"launch": true,
					},
				},
				{
					Name: "icu",
					Metadata: map[string]interface{}{
						"build": true,
					},
				},
			})

			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "icu",
				Metadata: map[string]interface{}{
					"build":  true,
					"launch": true,
				},
			}))
		})

	})
}
