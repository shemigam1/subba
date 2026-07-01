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

// TenantLookup resolves which tenant owns a given virtual account number.
// Implemented by a Postgres-backed type once that's wired up; kept as an
// interface here so this package doesn't need to know about pgx/sql.
type TenantLookup interface {
	TenantID(ctx context.Context, accountNumber string) (string, error)
}

// Publisher publishes a broker WEvent under a routing key.
type Publisher interface {
	Publish(ctx context.Context, routingKey string, event broker.Event) error
}

// CustomerLookup resolves which customer (if any) is tied to a virtual account.
type CustomerLookup interface {
	CustomerIDForVirtualAccount(ctx context.Context, accountNumber string) (string, error)
}

// SubscriptionLookup resolves which subscription owns a given customer.
type SubscriptionLookup interface {
	SubscriptionIDForCustomer(ctx context.Context, tenantID, customerID string) (string, error)
}

type Handler struct {
	WebhookSecret string
	Tenants       TenantLookup
	Customers     CustomerLookup
	Subscriptions SubscriptionLookup
	Publisher     Publisher
	Logger        zerolog.Logger
}

// ServeHTTP handles POST /webhooks/nomba.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Read the raw body — needed for signature verification before we
	// trust anything in it.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Warn().Err(err).Msg("webhook: failed to read body")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 2. Verify the signature before parsing further.
	signature := r.Header.Get("nomba-signature")
	timestamp := r.Header.Get("nomba-timestamp")
	if err := nomba.Verify(body, signature, timestamp, h.WebhookSecret); err != nil {
		h.Logger.Warn().Err(err).Msg("webhook: signature verification failed")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// 3. Parse the payload now that it's verified.
	var raw RawPayload
	if err := json.Unmarshal(body, &raw); err != nil {
		h.Logger.Warn().Err(err).Msg("webhook: failed to parse payload")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 4. Look up which tenant owns this virtual account.
	accountNumber := raw.Data.Transaction.AliasAccountNumber
	if accountNumber == "" {
		h.Logger.Warn().Str("event_type", raw.EventType).Msg("webhook: no virtual account number on payload")
		http.Error(w, "unrecognized payload shape", http.StatusUnprocessableEntity)
		return
	}
	tenantID, err := h.Tenants.TenantID(ctx, accountNumber)
	if err != nil {
		h.Logger.Error().Err(err).Str("account_number", accountNumber).Msg("webhook: tenant lookup failed")
		// Return 500 (not 400) — Nomba should retry, this is likely transient
		// (DB hiccup) or an event slightly ahead of account provisioning.
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// 5. Translate to our internal Event shape.
	event, rk, err := translate(raw, tenantID)
	if err != nil {
		h.Logger.Warn().Err(err).Msg("webhook: translate failed")
		http.Error(w, "unrecognized event type", http.StatusUnprocessableEntity)
		return
	}

	// 6. Look up the customer for this virtual account.
	customerID, err := h.Customers.CustomerIDForVirtualAccount(ctx, accountNumber)
	if err != nil {
		h.Logger.Error().Err(err).Str("account_number", accountNumber).Msg("webhook: customer lookup failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
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
