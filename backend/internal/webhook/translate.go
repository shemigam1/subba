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
	Kind       string  // internal routing key, e.g. "payment.succeeded"
	Amount     float64 // raw float from Nomba (in naira); ToBrokerEvent converts to kobo
	Reference  string  // merchantTxRef falls back to transactionId
	NombaTxID  string
	NombaReqID string
	OccurredAt time.Time
}

// routingKey maps a raw Nomba event type to the exchange routing key used by
// broker.DeclareTopology. The mapping is intentionally explicit — adding new
// Nomba event types should be a conscious decision that also updates topology.go.
func routingKey(rawEventType string) (string, error) {
	switch rawEventType {
	// ── Inbound bank-transfer credit (the "cardless renewal" moat) ─────────
	// These all arrive via the virtual-account webhook and represent money
	// landing in the customer's Nomba wallet. They trigger invoice creation
	// and, once confirmed, subscription advancement.
	case nomba.RawEventPaymentSuccess:
		return broker.RoutingKeyPaymentSucceeded, nil

	case nomba.RawEventPaymentFailed:
		return broker.RoutingKeyPaymentFailed, nil

	case nomba.RawEventPaymentReversal:
		return broker.RoutingKeyPaymentReversal, nil

	// ── Nomba-initiated payout events (tenant revenue split) ───────────────
	case nomba.RawEventPayoutSuccess, nomba.RawEventPayoutFailed,
		nomba.RawEventPayoutRefund:
		return broker.RoutingKeyPaymentSucceeded, nil

	// ── Scheduled subscription renewal (emitted by the scheduler, not Nomba) ─
	// The scheduler publishes directly to the exchange under this key; the
	// webhook path will never emit it, but routingKey is also called when
	// building subscription.renew events internally so we handle it here for
	// completeness.
	case broker.RoutingKeySubscriptionRenew:
		return broker.RoutingKeySubscriptionRenew, nil

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
// customerID and subscriptionID have already been resolved by the webhook handler.
func ToBrokerEvent(we Event, customerID, subscriptionID string) broker.Event {
	return broker.Event{
		RequestID:      we.NombaReqID,
		TenantID:       we.TenantID,
		EventType:      routingKeyOrKind(we.Kind),
		Amount:         int64(we.Amount * 100), // naira float → kobo int
		Currency:       "NGN",
		CustomerID:     customerID,
		SubscriptionID: subscriptionID,
		NombaReference: we.NombaTxID,
	}
}

// ToRenewalBrokerEvent builds a broker.Event for the subscription.renew routing
// path. This is called by the scheduler (not the webhook handler) when it finds a
// subscription whose period has elapsed and needs to be recharged. The references
// payload carries the plan ID so downstream handlers know which amount to charge.
func ToRenewalBrokerEvent(tenantID, customerID, subscriptionID, planID, requestID string) broker.Event {
	return broker.Event{
		RequestID:      requestID,
		TenantID:       tenantID,
		EventType:      broker.RoutingKeySubscriptionRenew,
		Amount:         0, // not yet known — the invoicing handler will resolve it from the plan
		Currency:       "NGN",
		CustomerID:     customerID,
		SubscriptionID: subscriptionID,
		NombaReference: "", // absent on subscription.renew; handlers must not assume it is present
		PlanID:         planID,
	}
}

// routingKeyOrKind returns the broker routing key for a given raw Nomba event
// type, falling back to the kind string if the mapping is unknown. This keeps
// EventType on the broker envelope aligned with the routing keys in topology.go.
func routingKeyOrKind(rawKind string) string {
	rk, err := routingKey(rawKind)
	if err != nil {
		return rawKind
	}
	return rk
}
