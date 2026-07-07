package nomba

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

// Verify checks a Nomba webhook's signature using standard HMAC on the raw HTTP body.
// Nomba documentation generally specifies HMAC-SHA256 (and sometimes SHA512).
// This function attempts to verify the signature using both algorithms, encoded as a hex string.
func Verify(rawBody []byte, signature, secret string) error {
	if signature == "" {
		return fmt.Errorf("missing nomba-signature header")
	}
	if secret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	secretBytes := []byte(secret)

	// Check SHA-256
	mac256 := hmac.New(sha256.New, secretBytes)
	mac256.Write(rawBody)
	expected256 := hex.EncodeToString(mac256.Sum(nil))
	if hmac.Equal([]byte(signature), []byte(expected256)) {
		return nil
	}

	// Check SHA-512 (fallback)
	mac512 := hmac.New(sha512.New, secretBytes)
	mac512.Write(rawBody)
	expected512 := hex.EncodeToString(mac512.Sum(nil))
	if hmac.Equal([]byte(signature), []byte(expected512)) {
		return nil
	}

	return fmt.Errorf("signature mismatch")
}
