package components

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

//go:generate faux --interface SignatureVerifier --output fakes/signature_verifier.go
type SignatureVerifier interface {
	Verify(signatureURL, targetURL string) error
}

func ConvertReleaseToDependency(release Release, signatureVerifier SignatureVerifier) (cargo.ConfigMetadataDependency, error) {
	var source, shasum512, asc ReleaseFile
	for _, f := range release.Files {
		if f.Name == fmt.Sprintf("icu4c-%s-src.tgz", strings.ReplaceAll(release.Version, ".", "_")) {
			source = f
		}

		if f.Name == fmt.Sprintf("icu4c-%s-src.tgz.asc", strings.ReplaceAll(release.Version, ".", "_")) {
			asc = f
		}

		if f.Name == "SHASUM512.txt" {
			shasum512 = f
		}
	}

	if (source == ReleaseFile{} || shasum512 == ReleaseFile{} || asc == ReleaseFile{}) {
		return cargo.ConfigMetadataDependency{}, fmt.Errorf("required files are missing from the release")
	}

	shasumResponse, err := http.Get(shasum512.URL)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}
	defer shasumResponse.Body.Close()

	b, err := io.ReadAll(shasumResponse.Body)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	r := regexp.MustCompile(fmt.Sprintf(`([0-9a-fA-F]+)  %s`, source.Name))

	submatch := r.FindStringSubmatch(string(b))
	if len(submatch) == 0 {
		return cargo.ConfigMetadataDependency{}, fmt.Errorf("unable to parse the shasum512 file")
	}
	checksum := submatch[1]

	purl := GeneratePURL("icu", release.Version, checksum, source.URL)

	licenses, err := GenerateLicenseInformation(source.URL)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	// Validate the artifact
	response, err := http.Get(source.URL)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}
	defer response.Body.Close()

	vr := cargo.NewValidatedReader(response.Body, fmt.Sprintf("sha512:%s", checksum))
	valid, err := vr.Valid()
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	if !valid {
		return cargo.ConfigMetadataDependency{}, fmt.Errorf("the given checksum of the source does not match with downloaded source")
	}

	err = signatureVerifier.Verify(asc.URL, source.URL)
	if err != nil {
		return cargo.ConfigMetadataDependency{}, err
	}

	return cargo.ConfigMetadataDependency{
		ID:             "icu",
		Name:           "ICU",
		Version:        release.Version,
		Source:         source.URL,
		SourceChecksum: fmt.Sprintf("sha512:%s", checksum),
		CPE:            fmt.Sprintf(`cpe:2.3:a:icu-project:international_components_for_unicode:%s:*:*:*:*:c\/c\+\+:*:*`, release.Version),
		PURL:           purl,
		Licenses:       licenses,
	}, nil
}
