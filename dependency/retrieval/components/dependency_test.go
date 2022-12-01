package components_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/icu/dependency/retrieval/components"
	"github.com/paketo-buildpacks/icu/dependency/retrieval/components/fakes"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

const (
	lFile = `The MIT License (MIT)

Copyright (c) .NET Foundation and Contributors

All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`
)

func testDependency(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("ConvertReleaseToDependeny", func() {
		var (
			server            *httptest.Server
			signatureVerifier *fakes.SignatureVerifier
		)

		it.Before(func() {
			buffer := bytes.NewBuffer(nil)
			gw := gzip.NewWriter(buffer)
			tw := tar.NewWriter(gw)

			Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err := tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			licenseFile := "some-dir/LICENSE.txt"
			Expect(tw.WriteHeader(&tar.Header{Name: licenseFile, Mode: 0755, Size: int64(len(lFile))})).To(Succeed())
			_, err = tw.Write([]byte(lFile))
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.Close()).To(Succeed())
			Expect(gw.Close()).To(Succeed())

			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodHead {
					http.Error(w, "NotFound", http.StatusNotFound)
					return
				}

				switch req.URL.Path {
				case "/source":
					w.WriteHeader(http.StatusOK)
					_, err := w.Write(buffer.Bytes())
					Expect(err).NotTo(HaveOccurred())

				case "/shasum512":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `d4bb1baed99674074f8af024dd159898eddaf4d71bc90f8d95b8448e96aac4b4e8358f755a516bfaf84baa34bf8657dc994459ef3bd72f54496b9ce2b0bd4636  icu4c-72_1-src.zip
a1aa65917e80e524c9b35466af83193001b1dfc030c5a084e02e2f71649a073e96382e9f561fb6378ace3f97402ebfb91beb815c18fea5c8136c3a9a04eff66c  icu4c-72_1-src.tgz`)

				case "/bad-shasum":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `d4bb1baed99674074f8af024dd159898eddaf4d71bc90f8d95b8448e96aac4b4e8358f755a516bfaf84baa34bf8657dc994459ef3bd72f54496b9ce2b0bd4636  icu4c-72_1-src.zip`)

				case "/wrong-shasum":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `aaaaaaaaaaaa  icu4c-72_1-src.tgz`)

				case "/bad-archive":
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte("\x66\x4C\x61\x43\x00\x00\x00\x22"))
					Expect(err).NotTo(HaveOccurred())

				default:
					t.Fatalf("unknown path: %s", req.URL.Path)
				}
			}))

			signatureVerifier = &fakes.SignatureVerifier{}
		})

		it("returns returns a cargo dependency generated from the given release", func() {
			dependency, err := components.ConvertReleaseToDependency(components.Release{
				SemVer:  semver.MustParse("72.1"),
				Version: "72.1",
				Files: []components.ReleaseFile{
					{
						Name: "icu4c-72_1-src.tgz",
						URL:  fmt.Sprintf("%s/source", server.URL),
					},
					{
						Name: "icu4c-72_1-src.tgz.asc",
						URL:  fmt.Sprintf("%s/source-asc", server.URL),
					},
					{
						Name: "SHASUM512.txt",
						URL:  fmt.Sprintf("%s/shasum512", server.URL),
					},
				},
			}, signatureVerifier)
			Expect(err).NotTo(HaveOccurred())

			Expect(dependency).To(Equal(cargo.ConfigMetadataDependency{
				Checksum:        "",
				CPE:             "cpe:2.3:a:icu-project:international_components_for_unicode:72.1:*:*:*:*:c\\/c\\+\\+:*:*",
				PURL:            fmt.Sprintf("pkg:generic/icu@72.1?checksum=a1aa65917e80e524c9b35466af83193001b1dfc030c5a084e02e2f71649a073e96382e9f561fb6378ace3f97402ebfb91beb815c18fea5c8136c3a9a04eff66c&download_url=%s/source", server.URL),
				ID:              "icu",
				Licenses:        []interface{}{"MIT", "MIT-0"},
				Name:            "ICU",
				SHA256:          "",
				Source:          fmt.Sprintf("%s/source", server.URL),
				SourceChecksum:  "sha512:a1aa65917e80e524c9b35466af83193001b1dfc030c5a084e02e2f71649a073e96382e9f561fb6378ace3f97402ebfb91beb815c18fea5c8136c3a9a04eff66c",
				SourceSHA256:    "",
				StripComponents: 0,
				URI:             "",
				Version:         "72.1",
			}))

			Expect(signatureVerifier.VerifyCall.Receives.SignatureURL).To(Equal(fmt.Sprintf("%s/source-asc", server.URL)))
			Expect(signatureVerifier.VerifyCall.Receives.TargetURL).To(Equal(fmt.Sprintf("%s/source", server.URL)))
		})

		context("failure cases", func() {
			context("when there are missing release files", func() {
				it("returns an error", func() {
					_, err := components.ConvertReleaseToDependency(components.Release{}, signatureVerifier)
					Expect(err).To(MatchError("required files are missing from the release"))
				})
			})

			context("when the shasum file get fails", func() {
				it("returns an error", func() {
					_, err := components.ConvertReleaseToDependency(components.Release{
						SemVer:  semver.MustParse("72.1"),
						Version: "72.1",
						Files: []components.ReleaseFile{
							{
								Name: "icu4c-72_1-src.tgz",
								URL:  fmt.Sprintf("%s/source", server.URL),
							},
							{
								Name: "icu4c-72_1-src.tgz.asc",
								URL:  fmt.Sprintf("%s/source-asc", server.URL),
							},
							{
								Name: "SHASUM512.txt",
								URL:  "not a valid url",
							},
						},
					}, signatureVerifier)
					Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
				})
			})

			context("when the shasum file cannot be parsed correctly", func() {
				it("returns an error", func() {
					_, err := components.ConvertReleaseToDependency(components.Release{
						SemVer:  semver.MustParse("72.1"),
						Version: "72.1",
						Files: []components.ReleaseFile{
							{
								Name: "icu4c-72_1-src.tgz",
								URL:  fmt.Sprintf("%s/source", server.URL),
							},
							{
								Name: "icu4c-72_1-src.tgz.asc",
								URL:  fmt.Sprintf("%s/source-asc", server.URL),
							},
							{
								Name: "SHASUM512.txt",
								URL:  fmt.Sprintf("%s/bad-shasum", server.URL),
							},
						},
					}, signatureVerifier)
					Expect(err).To(MatchError("unable to parse the shasum512 file"))
				})
			})

			context("when the artifact is not a supported archive type", func() {
				it("returns an error", func() {
					_, err := components.ConvertReleaseToDependency(components.Release{
						SemVer:  semver.MustParse("72.1"),
						Version: "72.1",
						Files: []components.ReleaseFile{
							{
								Name: "icu4c-72_1-src.tgz",
								URL:  fmt.Sprintf("%s/bad-archive", server.URL),
							},
							{
								Name: "icu4c-72_1-src.tgz.asc",
								URL:  fmt.Sprintf("%s/source-asc", server.URL),
							},
							{
								Name: "SHASUM512.txt",
								URL:  fmt.Sprintf("%s/shasum512", server.URL),
							},
						},
					}, signatureVerifier)
					Expect(err).To(MatchError(ContainSubstring("unsupported archive type")))
				})
			})

			context("when the checksum does not match", func() {
				it("returns an error", func() {
					_, err := components.ConvertReleaseToDependency(components.Release{
						SemVer:  semver.MustParse("72.1"),
						Version: "72.1",
						Files: []components.ReleaseFile{
							{
								Name: "icu4c-72_1-src.tgz",
								URL:  fmt.Sprintf("%s/source", server.URL),
							},
							{
								Name: "icu4c-72_1-src.tgz.asc",
								URL:  fmt.Sprintf("%s/source-asc", server.URL),
							},
							{
								Name: "SHASUM512.txt",
								URL:  fmt.Sprintf("%s/wrong-shasum", server.URL),
							},
						},
					}, signatureVerifier)
					Expect(err).To(MatchError("the given checksum of the source does not match with downloaded source"))
				})
			})

			context("when the verifier fails", func() {
				it.Before(func() {
					signatureVerifier.VerifyCall.Returns.Error = fmt.Errorf("verifier failed")
				})

				it("returns an error", func() {
					_, err := components.ConvertReleaseToDependency(components.Release{
						SemVer:  semver.MustParse("72.1"),
						Version: "72.1",
						Files: []components.ReleaseFile{
							{
								Name: "icu4c-72_1-src.tgz",
								URL:  fmt.Sprintf("%s/source", server.URL),
							},
							{
								Name: "icu4c-72_1-src.tgz.asc",
								URL:  fmt.Sprintf("%s/source-asc", server.URL),
							},
							{
								Name: "SHASUM512.txt",
								URL:  fmt.Sprintf("%s/shasum512", server.URL),
							},
						},
					}, signatureVerifier)
					Expect(err).To(MatchError("verifier failed"))
				})
			})
		})
	})
}
