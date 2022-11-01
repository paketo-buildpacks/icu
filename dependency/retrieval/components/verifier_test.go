package components_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paketo-buildpacks/icu/dependency/retrieval/components"
	"github.com/sclevine/spec"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"

	. "github.com/onsi/gomega"
)

func testVerifier(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect

		server   *httptest.Server
		verifier components.Verifier
	)

	it.Before(func() {
		buffer := bytes.NewBuffer(nil)
		gw := gzip.NewWriter(buffer)
		tw := tar.NewWriter(gw)

		licenseFile := "./LICENSE.txt"
		Expect(tw.WriteHeader(&tar.Header{Name: licenseFile, Mode: 0755, Size: int64(len(lFile))})).To(Succeed())
		_, err := tw.Write([]byte(lFile))
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.Close()).To(Succeed())
		Expect(gw.Close()).To(Succeed())

		entity, err := openpgp.NewEntity("", "", "", nil)
		Expect(err).NotTo(HaveOccurred())

		// A message reader needs to be created to prevent the cursor from moving
		// in the tar.gz buffer
		message := bytes.NewReader(buffer.Bytes())
		ascBuffer := bytes.NewBuffer(nil)
		err = openpgp.ArmoredDetachSign(ascBuffer, entity, message, nil)
		Expect(err).NotTo(HaveOccurred())

		primaryKey := bytes.NewBuffer(nil)

		err = entity.Serialize(primaryKey)
		Expect(err).NotTo(HaveOccurred())

		armoredPubKey := bytes.NewBuffer(nil)

		armoredKeyWriter, err := armor.Encode(armoredPubKey, "PGP PUBLIC KEY BLOCK", nil)
		Expect(err).NotTo(HaveOccurred())

		_, err = armoredKeyWriter.Write(primaryKey.Bytes())
		Expect(err).NotTo(HaveOccurred())

		armoredKeyWriter.Close()

		verifier = components.NewVerifier().WithPublicKey(armoredPubKey.String())

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

			case "/source-asc":
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(ascBuffer.Bytes())
				Expect(err).NotTo(HaveOccurred())

			case "/non-200":
				w.WriteHeader(http.StatusTeapot)

			default:
				t.Fatalf("unknown path: %s", req.URL.Path)
			}
		}))
	})

	context("Verify", func() {
		it("verifies the target with the asc file", func() {
			err := verifier.Verify(fmt.Sprintf("%s/source-asc", server.URL), fmt.Sprintf("%s/source", server.URL))
			Expect(err).NotTo(HaveOccurred())
		})

		context("failure cases", func() {
			context("when the target get failed", func() {
				it("returns an error", func() {
					err := verifier.Verify(fmt.Sprintf("%s/source-asc", server.URL), "not a valid url")
					Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
				})
			})

			context("when the target get is a non 200 status code", func() {
				it("returns an error", func() {
					err := verifier.Verify(fmt.Sprintf("%s/source-asc", server.URL), fmt.Sprintf("%s/non-200", server.URL))
					Expect(err).To(MatchError(fmt.Sprintf("received a non 200 status code from %s: status code 418 received", fmt.Sprintf("%s/non-200", server.URL))))
				})
			})

			context("when public key is not armored", func() {
				it.Before(func() {
					verifier = verifier.WithPublicKey("")
				})

				it("returns an error", func() {
					err := verifier.Verify(fmt.Sprintf("%s/source-asc", server.URL), fmt.Sprintf("%s/source", server.URL))
					Expect(err).To(MatchError(ContainSubstring("no armored data found")))
				})
			})

			context("when the signature get failed", func() {
				it("returns an error", func() {
					err := verifier.Verify("not a valid url", fmt.Sprintf("%s/source", server.URL))
					Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
				})
			})

			context("when the signature get is a non 200 status code", func() {
				it("returns an error", func() {
					err := verifier.Verify(fmt.Sprintf("%s/non-200", server.URL), fmt.Sprintf("%s/source", server.URL))
					Expect(err).To(MatchError(fmt.Sprintf("received a non 200 status code from %s: status code 418 received", fmt.Sprintf("%s/non-200", server.URL))))
				})
			})

			context("when the detached signature does not match the key", func() {
				it.Before(func() {
					entity, err := openpgp.NewEntity("", "", "", nil)
					Expect(err).NotTo(HaveOccurred())

					primaryKey := bytes.NewBuffer(nil)

					err = entity.Serialize(primaryKey)
					Expect(err).NotTo(HaveOccurred())

					armoredPubKey := bytes.NewBuffer(nil)

					armoredKeyWriter, err := armor.Encode(armoredPubKey, "PGP PUBLIC KEY BLOCK", nil)
					Expect(err).NotTo(HaveOccurred())

					_, err = armoredKeyWriter.Write(primaryKey.Bytes())
					Expect(err).NotTo(HaveOccurred())

					armoredKeyWriter.Close()

					verifier = verifier.WithPublicKey(armoredPubKey.String())
				})

				it("returns an error", func() {
					err := verifier.Verify(fmt.Sprintf("%s/source-asc", server.URL), fmt.Sprintf("%s/source", server.URL))
					Expect(err).To(MatchError(ContainSubstring("signature made by unknown entity")))
				})
			})
		})
	})
}
