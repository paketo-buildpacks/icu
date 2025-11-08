package components

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/paketo-buildpacks/libdependency/collections"
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

//go:generate faux --interface SignatureVerifier --output fakes/signature_verifier.go
type SignatureVerifier interface {
	Verify(signatureURL, targetURL string) error
}

type IcuReleaseFiles struct {
	Source    ReleaseFile
	Signature ReleaseFile
	Shasum512 ReleaseFile
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

type Generator struct {
	SignatureVerifier SignatureVerifier
	Targets           []PlatformStackTarget
}

func NewGenerator() Generator {
	return Generator{
		SignatureVerifier: NewVerifier(),
		Targets:           getSupportedPlatformStackTargets(),
	}
}

func (g Generator) WithVerifier(signatureVerifier SignatureVerifier) Generator {
	g.SignatureVerifier = signatureVerifier
	return g
}

func (g Generator) WithTarget(target PlatformStackTarget) Generator {
	g.Targets = []PlatformStackTarget{target}
	return g
}

func getSupportedPlatformStackTargets() []PlatformStackTarget {
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

func (g Generator) GenerateMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	icuVersion := versionFetcher.(IcuRelease)

	version := icuVersion.ReleaseVersion
	icuUrls, err := getUrls(icuVersion)
	if err != nil {
		return nil, err
	}

	checksum, err := getChecksum(icuUrls.Shasum512.URL, icuUrls.Source.Name)
	if err != nil {
		return nil, err
	}
	err = validateReleaseFiles(icuUrls, checksum, g.SignatureVerifier)
	if err != nil {
		return nil, err
	}

	cpe := fmt.Sprintf(`cpe:2.3:a:icu-project:international_components_for_unicode:%s:*:*:*:*:c\/c\+\+:*:*`, icuVersion.ReleaseVersion)
	purl := retrieve.GeneratePURL("icu", icuVersion.ReleaseVersion, checksum, icuUrls.Source.URL)

	return collections.TransformFuncWithError(g.Targets, func(platformTarget PlatformStackTarget) (versionology.Dependency, error) {
		fmt.Printf("Generating metadata for %s %s %s %s\n", platformTarget.OS, platformTarget.Arch, platformTarget.Target, version)
		configMetadataDependency := cargo.ConfigMetadataDependency{
			ID:             "icu",
			Name:           "ICU",
			Version:        version,
			Source:         icuUrls.Source.URL,
			SourceChecksum: fmt.Sprintf("sha512:%s", checksum),
			CPE:            cpe,
			PURL:           purl,
			Licenses:       []interface{}{"BSD-2-Clause", "BSD-3-Clause", "ICU", "Unicode-TOU"},
			Stacks:         platformTarget.Stacks,
			OS:             platformTarget.OS,
			Arch:           platformTarget.Arch,
		}

		return versionology.NewDependency(configMetadataDependency, platformTarget.Target)
	})
}

func getUrls(release IcuRelease) (IcuReleaseFiles, error) {
	var source, shasum512, asc ReleaseFile
	for _, f := range release.Files {
		if f.Name == fmt.Sprintf("icu4c-%s-src.tgz", strings.ReplaceAll(release.ReleaseVersion, ".", "_")) {
			source = f
		}
		if f.Name == fmt.Sprintf("icu4c-%s-sources.tgz", release.ReleaseVersion) {
			source = f
		}

		if f.Name == fmt.Sprintf("icu4c-%s-src.tgz.asc", strings.ReplaceAll(release.ReleaseVersion, ".", "_")) {
			asc = f
		}
		if f.Name == fmt.Sprintf("icu4c-%s-sources.tgz.asc", release.ReleaseVersion) {
			asc = f
		}

		if f.Name == "SHASUM512.txt" {
			shasum512 = f
		}
	}

	if (source == ReleaseFile{} || shasum512 == ReleaseFile{} || asc == ReleaseFile{}) {
		return IcuReleaseFiles{}, fmt.Errorf("required files are missing from the release")
	}

	return IcuReleaseFiles{
		Source:    source,
		Signature: asc,
		Shasum512: shasum512,
	}, nil
}

func getChecksum(checksumUrl string, fileName string) (string, error) {
	shasumResponse, err := http.Get(checksumUrl)
	if err != nil {
		return "", err
	}
	defer shasumResponse.Body.Close()

	b, err := io.ReadAll(shasumResponse.Body)
	if err != nil {
		return "", err
	}

	r := regexp.MustCompile(fmt.Sprintf(`([0-9a-fA-F]+)[\s\*]+%s`, fileName))

	submatch := r.FindStringSubmatch(string(b))
	if len(submatch) == 0 {
		return "", fmt.Errorf("unable to parse the SHASUM512 file")
	}
	checksum := submatch[1]

	return checksum, nil
}

func validateReleaseFiles(files IcuReleaseFiles, checksum string, signatureVerifier SignatureVerifier) error {
	// Validate the checksum
	response, err := http.Get(files.Source.URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	vr := cargo.NewValidatedReader(response.Body, fmt.Sprintf("sha512:%s", checksum))
	valid, err := vr.Valid()
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("the given checksum of the source does not match with downloaded source")
	}

	// Validate the signature
	err = signatureVerifier.Verify(files.Signature.URL, files.Source.URL)
	if err != nil {
		return err
	}

	return nil
}
