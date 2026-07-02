package nomba

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// signaturePayload mirrors only the fields Nomba includes in its signature
// calculation. Nomba signs a specific set of fields concatenated together —
// not the raw request body — so this struct exists purely to extract those
// fields. It is intentionally separate from the fuller webhook payload types
// used elsewhere for actual event processing.
type signaturePayload struct {
	EventType string `json:"event_type"`
	RequestID string `json:"requestId"`
	Data      struct {
		Merchant struct {
			UserID   string `json:"userId"`
			WalletID string `json:"walletId"`
		} `json:"merchant"`
		Transaction struct {
			TransactionID string `json:"transactionId"`
			Type          string `json:"type"`
			Time          string `json:"time"`
			ResponseCode  string `json:"responseCode"`
		} `json:"transaction"`
	} `json:"data"`
}

// Verify checks a Nomba webhook's signature. rawBody is the raw JSON request
// body; signature and timestamp come from the "nomba-signature" and
// "nomba-timestamp" request headers respectively; secret is your configured
// Nomba webhook signing key (cfg.NombaWebhookSecret).
//
// Returns nil if the signature is valid, or an error describing why it
// isn't — callers should treat any non-nil error as "reject this request."
func Verify(rawBody []byte, signature, timestamp, secret string) error {
	if signature == "" {
		return fmt.Errorf("missing nomba-signature header")
	}
	if timestamp == "" {
		return fmt.Errorf("missing nomba-timestamp header")
	}
	if secret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	var p signaturePayload
	if err := json.Unmarshal(rawBody, &p); err != nil {
		return fmt.Errorf("decode payload for signature check: %w", err)
	}

	hashingPayload := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s",
		p.EventType,
		p.RequestID,
		p.Data.Merchant.UserID,
		p.Data.Merchant.WalletID,
		p.Data.Transaction.TransactionID,
		p.Data.Transaction.Type,
		p.Data.Transaction.Time,
		p.Data.Transaction.ResponseCode,
	)
	message := hashingPayload + ":" + timestamp

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
