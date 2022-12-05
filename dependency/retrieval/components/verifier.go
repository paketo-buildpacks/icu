package components

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
)

type Verifier struct {
	publicKeyBlock string
	errorWriter    io.Writer
}

func NewVerifier() Verifier {
	return Verifier{
		publicKeyBlock: icuPublicKeyBlock,
		errorWriter:    os.Stderr,
	}
}

func (v Verifier) WithPublicKeyBlock(key string) Verifier {
	v.publicKeyBlock = key
	return v
}

func (v Verifier) WithErrorWriter(writer io.Writer) Verifier {
	v.errorWriter = writer
	return v
}

func (v Verifier) Verify(signatureURL, targetURL string) error {
	targetResponse, err := http.Get(targetURL)
	if err != nil {
		return err
	}
	defer targetResponse.Body.Close()

	if !(targetResponse.StatusCode >= 200 && targetResponse.StatusCode < 300) {
		return fmt.Errorf("received a non 200 status code from %s: status code %d received", targetURL, targetResponse.StatusCode)
	}

	signatureResponse, err := http.Get(signatureURL)
	if err != nil {
		return err
	}
	defer signatureResponse.Body.Close()

	if !(signatureResponse.StatusCode >= 200 && signatureResponse.StatusCode < 300) {
		return fmt.Errorf("received a non 200 status code from %s: status code %d received", signatureURL, signatureResponse.StatusCode)
	}

	// Need to get all of the bytes to allow for the file to be read multiple times
	signatureBytes, err := io.ReadAll(signatureResponse.Body)
	if err != nil {
		return err
	}

	keys := parseKeys(v.publicKeyBlock)
	for _, key := range keys {
		keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(key))
		if err != nil {
			fmt.Fprintf(v.errorWriter, "failed to read armored key: %s\n", err.Error())
			continue
		}

		_, err = openpgp.CheckArmoredDetachedSignature(keyring, targetResponse.Body, bytes.NewReader(signatureBytes))
		if err != nil {
			fmt.Fprintf(v.errorWriter, "failed to check signature: %s\n", err.Error())
			continue
		}

		// Valid and matching PGP was found
		return nil
	}

	return errors.New("no valid pgp keys provided")
}

func parseKeys(block string) []string {
	var keys []string
	var currentKey string
	inKey := false

	for _, line := range strings.Split(string(block), "\n") {
		if line == "-----BEGIN PGP PUBLIC KEY BLOCK-----" {
			currentKey = line + "\n"
			inKey = true
		} else if line == "-----END PGP PUBLIC KEY BLOCK-----" {
			currentKey = currentKey + line
			keys = append(keys, currentKey)
			inKey = false
		} else if inKey {
			currentKey = currentKey + line + "\n"
		}
	}

	return keys
}
