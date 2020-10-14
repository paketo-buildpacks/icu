package icu_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitIcu(t *testing.T) {
	suite := spec.New("icu", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("ICULayerArranger", testICULayerArranger)
	suite("PlanEntryResolver", testPlanEntryResolver)
	suite.Run(t)
}
