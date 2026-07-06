package nomba

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// VerifyParams holds the individual fields extracted from the parsed webhook
// payload and headers that Nomba uses to construct the HMAC signature.
//
// Nomba does NOT sign the raw HTTP body. Instead, it concatenates these
// specific fields in a strict order separated by colons, appends the
// nomba-timestamp header, then HMAC-SHA256 + Base64 encodes the result.
type VerifyParams struct {
	EventType       string // payload.event_type
	RequestID       string // payload.requestId
	UserID          string // payload.data.merchant.userId
	WalletID        string // payload.data.merchant.walletId
	TransactionID   string // payload.data.transaction.transactionId
	TransactionType string // payload.data.transaction.type
	TransactionTime string // payload.data.transaction.time
	ResponseCode    string // payload.data.transaction.responseCode (empty string if "null")
	Timestamp       string // nomba-timestamp HTTP header
}

// Verify checks a Nomba webhook's signature using the proprietary algorithm
// documented at developer.nomba.com.
//
// The signature is constructed by:
//  1. Concatenating 8 payload fields + the nomba-timestamp header with ":" separators
//  2. Computing HMAC-SHA256 of that string using the webhook secret key
//  3. Base64-encoding the resulting hash
//
// Returns nil if the signature is valid.
func Verify(params VerifyParams, signature, secret string) error {
	if signature == "" {
		return fmt.Errorf("missing nomba-signature header")
	}
	if secret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	// Nomba treats "null" response codes as empty strings.
	responseCode := params.ResponseCode
	if responseCode == "null" {
		responseCode = ""
	}

	// Build the exact hashing payload Nomba uses.
	// Format: eventType:requestId:userId:walletId:transactionId:type:time:responseCode:timestamp
	hashingPayload := fmt.Sprintf(
		"%s:%s:%s:%s:%s:%s:%s:%s:%s",
		params.EventType,
		params.RequestID,
		params.UserID,
		params.WalletID,
		params.TransactionID,
		params.TransactionType,
		params.TransactionTime,
		responseCode,
		params.Timestamp,
	)

	// HMAC-SHA256 → Base64
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(hashingPayload))
	expected := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
