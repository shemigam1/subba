package webhook

import (
	"fmt"
	"time"

	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/nomba"
)

// Event is our internal representation of a payment event, decoupled from
// Nomba's wire format. Everything downstream (invoicing, payouts,
// subscription state) consumes this shape, not Nomba's raw payload — so if
// Nomba changes their webhook format, only translate.go needs to change.
type Event struct {
	TenantID   string
	Kind       string // internal event kind, e.g. "payment.success"
	Amount     float64
	Reference  string // merchantTxRef, falls back to transactionId
	NombaTxID  string
	NombaReqID string
	OccurredAt time.Time
}

// routingKey maps a raw Nomba event type to the exchange routing key used by
// broker.DeclareTopology.
func routingKey(rawEventType string) (string, error) {
	switch rawEventType {
	case nomba.RawEventPaymentSuccess, nomba.RawEventPaymentFailed, nomba.RawEventPaymentReversal,
		nomba.RawEventPayoutSuccess, nomba.RawEventPayoutFailed, nomba.RawEventPayoutRefund:
		return broker.RoutingKeyPaymentSucceeded, nil
	default:
		return "", fmt.Errorf("unrecognized event type: %s", rawEventType)
	}
}

// translate converts a raw Nomba payload plus a resolved tenant ID into our
// internal Event, and returns the routing key it should publish under.
func translate(raw RawPayload, tenantID string) (Event, string, error) {
	rk, err := routingKey(raw.EventType)
	if err != nil {
		return Event{}, "", err
	}

	ref := raw.Data.Transaction.MerchantTxRef
	if ref == "" {
		ref = raw.Data.Transaction.TransactionID
	}

	occurredAt, err := time.Parse(time.RFC3339, raw.Data.Transaction.Time)
	if err != nil {
		occurredAt = time.Now().UTC() // fallback if Nomba's timestamp is missing/malformed
	}

	event := Event{
		TenantID:   tenantID,
		Kind:       raw.EventType,
		Amount:     raw.Data.Transaction.Amount,
		Reference:  ref,
		NombaTxID:  raw.Data.Transaction.TransactionID,
		NombaReqID: raw.RequestID,
		OccurredAt: occurredAt,
	}
	return event, rk, nil
}

// ToBrokerEvent converts a webhook.Event to a broker.Event for publishing.
func ToBrokerEvent(we Event, customerID, subscriptionID string) broker.Event {
	return broker.Event{
		RequestID:      we.NombaReqID,
		TenantID:       we.TenantID,
		EventType:      we.Kind,
		Amount:         int64(we.Amount * 100), // float kobo → int kobo
		Currency:       "NGN",
		CustomerID:     customerID,
		SubscriptionID: subscriptionID,
		NombaReference: we.NombaTxID,
	}
}
