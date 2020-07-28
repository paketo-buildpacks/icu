package integration_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	buildpack          string
	buildPlanBuildpack string
	// offlineBuildpack   string
	buildpackInfo struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}
	config struct {
		BuildPlan string `json:"build-plan"`
	}
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.DecodeReader(file, &buildpackInfo)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	buildpack, err = Package(root, "1.2.3", false)
	Expect(err).NotTo(HaveOccurred())

	// offlineBuildpack, err = buildpackStore.Get.
	// 	WithOfflineDependencies().
	// 	WithVersion("1.2.3").
	// 	Execute(root)
	// Expect(err).NotTo(HaveOccurred())

	buildPlanBuildpack, err = buildpackStore.Get.
		Execute(config.BuildPlan)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault)
	//	suite("Offline", testOffline)
	suite.Run(t)
}

func Package(root, version string, cached bool) (string, error) {
	var cmd *exec.Cmd

	dir, err := filepath.Abs("./..")
	if err != nil {
		return "", err
	}

	bpPath := filepath.Join(root, "artifact")
	if cached {
		cmd = exec.Command(filepath.Join(dir, ".bin", "packager"), "--archive", "--version", version, fmt.Sprintf("%s-cached", bpPath))
	} else {
		cmd = exec.Command(filepath.Join(dir, ".bin", "packager"), "--archive", "--uncached", "--version", version, bpPath)
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("PACKAGE_DIR=%s", bpPath))
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	if cached {
		return fmt.Sprintf("%s-cached.tgz", bpPath), nil
	}

	return fmt.Sprintf("%s.tgz", bpPath), nil
}
