package nomba

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	secret := "test_secret"
	params := VerifyParams{
		EventType:       "payment_success",
		RequestID:       "req_123456",
		UserID:          "user_001",
		WalletID:        "wallet_001",
		TransactionID:   "txn_001",
		TransactionType: "vact_transfer",
		TransactionTime: "2026-07-06T09:00:00Z",
		ResponseCode:    "",
		Timestamp:       "2026-07-06T09:00:05Z",
	}

	// Reconstruct the expected hash logic
	hashingPayload := fmt.Sprintf(
		"%s:%s:%s:%s:%s:%s:%s:%s:%s",
		params.EventType,
		params.RequestID,
		params.UserID,
		params.WalletID,
		params.TransactionID,
		params.TransactionType,
		params.TransactionTime,
		params.ResponseCode,
		params.Timestamp,
	)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(hashingPayload))
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	err := Verify(params, expectedSignature, secret)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = Verify(params, "invalid_signature", secret)
	if err == nil {
		t.Fatal("Expected error for invalid signature, got nil")
	}
}
