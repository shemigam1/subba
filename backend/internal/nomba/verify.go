package nomba

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Verify checks a Nomba webhook's signature. rawBody is the raw JSON request
// body; signature comes from the "nomba-signature" request header;
// secret is your configured Nomba webhook signing key (cfg.NombaWebhookSecret).
//
// Returns nil if the signature is valid, or an error describing why it
// isn't — callers should treat any non-nil error as "reject this request."
func Verify(rawBody []byte, signature, secret string) error {
	if signature == "" {
		return fmt.Errorf("missing nomba-signature header")
	}
	if secret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	if expected != signature {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
