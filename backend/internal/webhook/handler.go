package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/nomba"
)

// Publisher publishes a broker Event under a routing key.
type Publisher interface {
	Publish(ctx context.Context, routingKey string, event broker.Event) error
}

// SubscriptionLookup resolves which subscription owns a given customer.
type SubscriptionLookup interface {
	SubscriptionIDForCustomer(ctx context.Context, tenantID, customerID string) (string, error)
}

type Handler struct {
	WebhookSecret string
	Subscriptions SubscriptionLookup
	Publisher     Publisher
	Logger        zerolog.Logger
}

// ServeHTTP handles POST /webhooks/nomba.
//
// IMPORTANT: Nomba uses a proprietary HMAC signature algorithm. They do NOT
// sign the raw HTTP body. Instead, they extract 8 specific fields from the
// parsed JSON, concatenate them with colons, append the nomba-timestamp
// header value, then HMAC-SHA256 + Base64 encode the result.
//
// This means we MUST parse the JSON body BEFORE we can verify the signature.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Read the raw body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Warn().Err(err).Msg("webhook: failed to read body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 2. Parse the payload FIRST — Nomba's signature verification requires
	//    extracting specific fields from the parsed payload, not hashing the
	//    raw body.
	var raw RawPayload
	if err := json.Unmarshal(body, &raw); err != nil {
		h.Logger.Warn().Err(err).Msg("webhook: failed to parse payload")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 3. Extract the nomba-timestamp header and the signature.
	//    HTTP header names are case-insensitive; Go canonicalizes them, but
	//    Nomba sends lowercase keys so we use the canonical getter.
	signature := r.Header.Get("nomba-signature")
	timestamp := r.Header.Get("nomba-timestamp")

	// 4. Verify the signature using Nomba's proprietary algorithm:
	//    HMAC-SHA256(eventType:requestId:userId:walletId:transactionId:type:time:responseCode:timestamp)
	//    then Base64 encode.
	verifyParams := nomba.VerifyParams{
		EventType:       raw.EventType,
		RequestID:       raw.RequestID,
		UserID:          raw.Data.Merchant.UserID,
		WalletID:        raw.Data.Merchant.WalletID,
		TransactionID:   raw.Data.Transaction.TransactionID,
		TransactionType: raw.Data.Transaction.Type,
		TransactionTime: raw.Data.Transaction.Time,
		ResponseCode:    raw.Data.Transaction.ResponseCode,
		Timestamp:       timestamp,
	}

	if err := nomba.Verify(verifyParams, signature, h.WebhookSecret); err != nil {
		h.Logger.Warn().Err(err).Msg("webhook: signature verification failed")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// 5. Parse the TenantID and CustomerID from the custom AliasAccountReference tag.
	// As per the Nomba Slack advisory, we tag virtual accounts with "{tenant_id}:{customer_id}".
	ref := raw.Data.Transaction.AliasAccountReference
	if ref == "" {
		h.Logger.Warn().Str("event_type", raw.EventType).Msg("webhook: missing aliasAccountReference on payload")
		http.Error(w, "unrecognized payload shape", http.StatusUnprocessableEntity)
		return
	}

	// Format is "tenantID:customerID"
	var tenantID, customerID string
	for i, c := range ref {
		if c == ':' {
			tenantID = ref[:i]
			customerID = ref[i+1:]
			break
		}
	}
	if tenantID == "" || customerID == "" {
		h.Logger.Warn().Str("ref", ref).Msg("webhook: invalid aliasAccountReference format")
		http.Error(w, "invalid account reference", http.StatusUnprocessableEntity)
		return
	}

	// 6. Translate to our internal Event shape.
	event, rk, err := translate(raw, tenantID)
	if err != nil {
		h.Logger.Warn().Err(err).Msg("webhook: translate failed")
		http.Error(w, "unrecognized event type", http.StatusUnprocessableEntity)
		return
	}

	// 7. Look up the subscription for this customer.
	subscriptionID, err := h.Subscriptions.SubscriptionIDForCustomer(ctx, tenantID, customerID)
	if err != nil {
		h.Logger.Error().Err(err).
			Str("tenant_id", tenantID).
			Str("customer_id", customerID).
			Msg("webhook: subscription lookup failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// 8. Convert to broker.Event and publish — only return 200 once the broker has confirmed receipt.
	brokerEvent := ToBrokerEvent(event, customerID, subscriptionID)
	if err := h.Publisher.Publish(ctx, rk, brokerEvent); err != nil {
		h.Logger.Error().Err(err).
			Str("routing_key", rk).
			Str("tenant_id", tenantID).
			Msg("webhook: publish failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.Logger.Info().
		Str("tenant_id", tenantID).
		Str("event_kind", event.Kind).
		Str("reference", event.Reference).
		Msg("webhook: event published")

	w.WriteHeader(http.StatusOK)
}
