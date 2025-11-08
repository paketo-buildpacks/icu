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
	stacks []string
	target string
	os     string
	arch   string
}

func getSupportedPlatformStackTargets() []PlatformStackTarget {
	var platformStackTargets []PlatformStackTarget

	for os, architectures := range supportedPlatforms {
		for _, arch := range architectures {
			for _, pair := range supportedStacks {
				platformStackTargets = append(platformStackTargets, PlatformStackTarget{
					stacks: pair.stacks,
					target: pair.target,
					os:     os,
					arch:   arch,
				})
			}
		}
	}

	return platformStackTargets
}

func WriteOutput(path string, dependencies []cargo.ConfigMetadataDependency) error {
	var output []OutputDependency
	for _, dependency := range dependencies {
		for _, platformTarget := range getSupportedPlatformStackTargets() {
			dependency.Stacks = platformTarget.stacks
			dependency.OS = platformTarget.os
			dependency.Arch = platformTarget.arch
			output = append(output, OutputDependency{
				ConfigMetadataDependency: dependency,
				Target:                   platformTarget.target,
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
