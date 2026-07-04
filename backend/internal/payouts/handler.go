// Package payouts was originally intended to implement a payouts consumer handler.
//
// # OBSOLETE - DO NOT IMPLEMENT
//
// This worker is permanently obsolete and intentionally disabled.
// According to Nomba's Instant Settlement architecture, when a virtual account
// is created and scoped to a sub-account ID, Nomba automatically settles funds
// directly into the sub-account upon payment.
//
// DO NOT attempt to implement a manual `Transfer` here. If you initiate a manual
// transfer from the platform's balance to the sub-account, you will double-pay
// the tenant (once automatically by Nomba, and once manually by this worker).
package payouts

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/idempotency"
)

// NewHandler returns the idempotency.HandlerFunc for the payouts consumer.
// The payouts worker is permanently disabled.
func NewHandler(log zerolog.Logger) idempotency.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, event broker.Event) error {
		log.Debug().Str("event_type", event.EventType).Msg("payouts: worker intentionally obsolete due to instant settlement, skipping event")
		return nil
	}
}
