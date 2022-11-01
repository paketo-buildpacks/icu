package components

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/openpgp"
)

// This key was taken from https://github.com/unicode-org/icu/blob/main/KEYS on 2022-10-27
const (
	icuPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

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
wkUpoDU7dMYRZrtm3BIMdgntWYSs
=52fO
-----END PGP PUBLIC KEY BLOCK-----`
)

type Verifier struct {
	publicKey string
}

func NewVerifier() Verifier {
	return Verifier{
		publicKey: icuPublicKey,
	}
}

func (v Verifier) WithPublicKey(key string) Verifier {
	v.publicKey = key
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

	keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(v.publicKey))
	if err != nil {
		return err
	}

	signatureResponse, err := http.Get(signatureURL)
	if err != nil {
		return err
	}
	defer signatureResponse.Body.Close()

	if !(signatureResponse.StatusCode >= 200 && signatureResponse.StatusCode < 300) {
		return fmt.Errorf("received a non 200 status code from %s: status code %d received", signatureURL, signatureResponse.StatusCode)
	}

	_, err = openpgp.CheckArmoredDetachedSignature(keyring, targetResponse.Body, signatureResponse.Body)
	if err != nil {
		return err
	}

	return nil
}
