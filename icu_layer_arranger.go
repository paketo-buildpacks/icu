package icu

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/fs"
)

type ICULayerArranger struct{}

func NewICULayerArranger() ICULayerArranger {
	return ICULayerArranger{}
}

func (a ICULayerArranger) Arrange(path string) error {
	files, err := filepath.Glob(filepath.Join(path, "usr", "local", "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		err = fs.Copy(file, filepath.Join(path, filepath.Base(file)))
		if err != nil {
			return err
		}
	}

	err = os.RemoveAll(filepath.Join(path, "usr"))
	if err != nil {
		return err
	}

	return nil
}
