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

func WriteOutput(path string, dependencies []OutputDependency) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(dependencies)
	if err != nil {
		return err
	}

	return nil
}
