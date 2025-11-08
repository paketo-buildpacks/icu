package components

import (
	"encoding/json"
	"os"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type OutputDependency struct {
	cargo.ConfigMetadataDependency
	Target string `json:"target"`
}

type StackAndTargetPair struct {
	stacks []string
	target string
}

var supportedStacks = []StackAndTargetPair{
	{stacks: []string{"io.buildpacks.stacks.jammy"}, target: "jammy"},
	{stacks: []string{"io.buildpacks.stacks.noble"}, target: "noble"},
}

var supportedPlatforms = map[string][]string{
	"linux": {"amd64", "arm64"},
}

type PlatformStackTarget struct {
	Stacks []string
	Target string
	OS     string
	Arch   string
}

func GetSupportedPlatformStackTargets() []PlatformStackTarget {
	var platformStackTargets []PlatformStackTarget

	for os, architectures := range supportedPlatforms {
		for _, arch := range architectures {
			for _, pair := range supportedStacks {
				platformStackTargets = append(platformStackTargets, PlatformStackTarget{
					Stacks: pair.stacks,
					Target: pair.target,
					OS:     os,
					Arch:   arch,
				})
			}
		}
	}

	return platformStackTargets
}

func WriteOutput(path string, dependencies []cargo.ConfigMetadataDependency, platformTargets []PlatformStackTarget) error {
	var output []OutputDependency
	for _, dependency := range dependencies {
		for _, platformTarget := range platformTargets {
			dependency.Stacks = platformTarget.Stacks
			dependency.OS = platformTarget.OS
			dependency.Arch = platformTarget.Arch
			output = append(output, OutputDependency{
				ConfigMetadataDependency: dependency,
				Target:                   platformTarget.Target,
			})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(output)
	if err != nil {
		return err
	}

	return nil
}
