// Package payouts implements the payouts consumer handler.
//
// # Responsibility
//
// When a payment.succeeded event is received, the payouts handler initiates a
// bank transfer from the platform's Nomba balance to the tenant's sub-account.
// The transfer amount is the event.Amount (the exact kobo received from the customer).
//
// # Two-phase flow
//
//  1. Write a 'pending' payout record to the invoices table (status='open',
//     nomba_reference=nil). This is the "pending payout" signal.
//  2. Call Nomba Transfer with event.RequestID as the merchantTxRef (Nomba's
//     idempotency key) so retries are safe.
//  3. Update the record based on the Nomba response:
//     - success → MarkInvoicePaid with the Nomba transaction ID
//     - failure → leave the invoice open (the wrapper marks processed_events
//     as 'failed'; the retry queue will redeliver)
//
// # Idempotency boundary
//
// All deduplication is the wrapper's job. The handler only reads the tx passed
// to it and writes through that tx. It never touches processed_events.
//
// # Note on payout invoice
//
// We reuse the invoices table to record payout disbursements. A payout invoice
// has subscription_id=NULL and its description line item reads "Payout to tenant".
// This keeps the accounting schema simple — one table, immutable rows.
package payouts

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/idempotency"
)

// NewHandler returns the idempotency.HandlerFunc for the payouts consumer.
// The payouts worker is currently disabled under Architecture A.
func NewHandler(log zerolog.Logger) idempotency.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, event broker.Event) error {
		log.Debug().Str("event_type", event.EventType).Msg("payouts: worker disabled, skipping event")
		return nil
	}
}
