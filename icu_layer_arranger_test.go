package icu_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	icu "github.com/paketo-buildpacks/icu"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testICULayerArranger(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layerDir string

		arranger icu.ICULayerArranger
	)

	it.Before(func() {
		var err error
		layerDir, err = os.MkdirTemp("", "layer")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(layerDir, "usr", "local", "bin"), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(layerDir, "usr", "local", "some-file"), nil, os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(layerDir, "usr", "local", "bin", "some-other-file"), nil, os.ModePerm)).To(Succeed())

		arranger = icu.NewICULayerArranger()
	})

	it.After(func() {
		Expect(os.RemoveAll(layerDir)).To(Succeed())
	})

	context("Arrange", func() {
		it("extracts all contents out of usr/local and puts them in the root of the layer then deletes usr", func() {
			err := arranger.Arrange(layerDir)
			Expect(err).NotTo(HaveOccurred())

			var files []string
			err = filepath.Walk(layerDir, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					rel, err := filepath.Rel(layerDir, path)
					if err != nil {
						log.Fatal(err)
					}

					files = append(files, rel)
					return nil
				}
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(files).To(ConsistOf([]string{
				filepath.Join("bin/some-other-file"),
				filepath.Join("some-file"),
			}))

			Expect(filepath.Join(layerDir, "usr")).NotTo(BeADirectory())
		})

		context("error cases", func() {
			context("when the set of files in usr/local can't be gotten", func() {
				it("returns the error", func() {
					err := arranger.Arrange("\\")
					Expect(err).To(MatchError(ContainSubstring("syntax error in pattern")))
				})
			})

			context("when copying fails", func() {
				it.Before(func() {
					Expect(os.MkdirAll(filepath.Join(layerDir, "bin"), os.ModePerm)).To(Succeed())
					Expect(os.WriteFile(filepath.Join(layerDir, "bin", "some-other-file"), nil, os.ModePerm)).To(Succeed())
				})
				it("returns an error", func() {
					err := arranger.Arrange(layerDir)
					Expect(err).To(MatchError(ContainSubstring("failed to copy: destination exists")))
				})
			})
		})
	})

}
