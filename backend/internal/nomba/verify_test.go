package nomba

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	secret := "test_secret"
	rawBody := []byte(`{"event_type":"payment_success","data":{"transaction":{"responseCode":"00"}}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	err := Verify(rawBody, expectedSignature, secret)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = Verify(rawBody, "invalid_signature", secret)
	if err == nil {
		t.Fatal("Expected error for invalid signature, got nil")
	}
}
