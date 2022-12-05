package components

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type Release struct {
	SemVer  *semver.Version
	Version string
	Files   []ReleaseFile
}

type ReleaseFile struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type Fetcher struct {
	api string
}

func NewFetcher() Fetcher {
	return Fetcher{
		api: "https://api.github.com",
	}
}

func (f Fetcher) WithAPI(uri string) Fetcher {
	f.api = uri
	return f
}

func (f Fetcher) Get() ([]Release, error) {
	page := 1
	var releases []Release
	for {
		resp, err := http.Get(fmt.Sprintf("%s/repos/unicode-org/icu/releases?per_page=100&page=%d", f.api, page))
		if err != nil {
			return nil, err
		}

		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			return nil, fmt.Errorf("received a non 200 status code: status code %d received", resp.StatusCode)
		}

		var releaseResponse []struct {
			TagName    string        `json:"tag_name"`
			Name       string        `json:"name"`
			Draft      bool          `json:"draft"`
			Prerelease bool          `json:"prerelease"`
			Assets     []ReleaseFile `json:"assets"`
		}
		err = json.NewDecoder(resp.Body).Decode(&releaseResponse)
		if err != nil {
			return nil, err
		}

		if len(releaseResponse) == 0 {
			break
		}

		page++

		for _, release := range releaseResponse {
			if release.Draft || release.Prerelease {
				continue
			}

			var r Release
			r.Version = strings.TrimPrefix(release.Name, "ICU ") // The space is important
			r.SemVer, err = semver.NewVersion(r.Version)
			if err != nil {
				return nil, fmt.Errorf("%w: the following version string could not be parsed %q", err, r.Version)
			}
			r.Files = release.Assets

			releases = append(releases, r)
		}
	}

	return releases, nil
}
