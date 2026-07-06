// Package substate implements the subscription-state consumer handler.
//
// # Responsibility
//
//   - payment.succeeded → advance the billing period + set status=active
//   - subscription.renew → initiate a tokenized-card charge via Nomba;
//     on success, advance the period; on failure, set status=past_due
//
// # Transaction contract
//
// The handler receives a live Postgres transaction (tx) from the idempotency
// wrapper. All DB writes go through that tx. The wrapper commits or rolls
// back. The handler never touches processed_events directly.
//
// # Billing period arithmetic
//
// AddDate is used for calendar-aware month/year addition (handles month-end
// rollover correctly — Go's time.AddDate normalises Feb 31 → Mar 3, etc.).
package substate

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/idempotency"
	"github.com/shamigam1/subba/internal/nomba"
	"github.com/shamigam1/subba/internal/store/db"
)

// NewHandler returns the idempotency.HandlerFunc for the subscription_state consumer.
func NewHandler(nombaClient *nomba.Client, log zerolog.Logger) idempotency.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, event broker.Event) error {
		q := db.New(tx)
		switch event.EventType {
		case broker.RoutingKeyPaymentSucceeded:
			return handlePaymentSucceeded(ctx, q, event, log)
		case broker.RoutingKeyPaymentFailed, broker.RoutingKeyPaymentReversal:
			return handlePaymentFailure(ctx, q, event, log)
		case broker.RoutingKeySubscriptionRenew:
			return handleSubscriptionRenew(ctx, q, nombaClient, event, log)
		default:
			return fmt.Errorf("substate: unhandled event type %q", event.EventType)
		}
	}
}

// ── payment.succeeded ─────────────────────────────────────────────────────────

// handlePaymentSucceeded advances the billing period and sets status=active.
// This is called when a Nomba virtual-account credit is confirmed — the
// customer has paid for the next cycle via bank transfer (the "cardless moat").
func handlePaymentSucceeded(ctx context.Context, q *db.Queries, event broker.Event, log zerolog.Logger) error {
	subID, err := uuid.Parse(event.SubscriptionID)
	if err != nil {
		return fmt.Errorf("substate: parse subscription_id: %w", err)
	}

	sub, err := q.GetSubscription(ctx, subID)
	if err != nil {
		return fmt.Errorf("substate: get subscription: %w", err)
	}

	plan, err := q.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("substate: get plan for subscription: %w", err)
	}

	newStart, newEnd := nextPeriod(sub, plan)

	updated, err := q.AdvanceSubscriptionPeriod(ctx, db.AdvanceSubscriptionPeriodParams{
		ID:             subID,
		NewPeriodStart: pgtype.Timestamptz{Time: newStart, Valid: true},
		NewPeriodEnd:   pgtype.Timestamptz{Time: newEnd, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("substate: advance subscription period: %w", err)
	}

	log.Info().
		Str("subscription_id", updated.ID.String()).
		Str("status", updated.Status).
		Time("period_start", newStart).
		Time("period_end", newEnd).
		Msg("substate: billing period advanced — subscription active")
	return nil
}

// ── payment.failed & payment.reversal ─────────────────────────────────────────

// handlePaymentFailure handles payment reversals and failures by locking
// the subscription to past_due, requiring manual or scheduler retry intervention.
func handlePaymentFailure(ctx context.Context, q *db.Queries, event broker.Event, log zerolog.Logger) error {
	subID, err := uuid.Parse(event.SubscriptionID)
	if err != nil {
		return fmt.Errorf("substate: parse subscription_id for failure: %w", err)
	}

	log.Warn().
		Str("subscription_id", event.SubscriptionID).
		Str("event_type", event.EventType).
		Msg("substate: payment failed or reversed — marking subscription past_due")

	return setStatus(ctx, q, subID, "past_due", log)
}

// ── subscription.renew ────────────────────────────────────────────────────────

// handleSubscriptionRenew is called by the scheduler on the recurring billing
// cycle. It charges the customer's saved tokenized card via Nomba, then:
//   - on success: advance the period and set status=active
//   - on failure: set status=past_due (the scheduler will retry later)
func handleSubscriptionRenew(
	ctx context.Context,
	q *db.Queries,
	nombaClient *nomba.Client,
	event broker.Event,
	log zerolog.Logger,
) error {
	subID, err := uuid.Parse(event.SubscriptionID)
	if err != nil {
		return fmt.Errorf("substate renew: parse subscription_id: %w", err)
	}
	customerID, err := uuid.Parse(event.CustomerID)
	if err != nil {
		return fmt.Errorf("substate renew: parse customer_id: %w", err)
	}

	sub, err := q.GetSubscription(ctx, subID)
	if err != nil {
		return fmt.Errorf("substate renew: get subscription: %w", err)
	}

	plan, err := q.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("substate renew: get plan: %w", err)
	}

	// Fetch the customer's saved token key.
	customer, err := q.GetCustomer(ctx, customerID)
	if err != nil {
		return fmt.Errorf("substate renew: get customer: %w", err)
	}

	tenant, err := q.GetTenantByID(ctx, sub.TenantID)
	if err != nil {
		return fmt.Errorf("substate renew: get tenant: %w", err)
	}

	if customer.NombaTokenKey == nil || *customer.NombaTokenKey == "" {
		// No card token — can't charge; mark past_due and let the scheduler
		// prompt the customer to re-add their card.
		log.Warn().
			Str("subscription_id", event.SubscriptionID).
			Str("customer_id", event.CustomerID).
			Msg("substate renew: no card token — marking past_due")
		return setStatus(ctx, q, subID, "past_due", log)
	}

	// ── Charge the tokenized card ─────────────────────────────────────────────
	// merchantTxRef = "renew:{requestID}" — unique per attempt, safe to retry.
	merchantTxRef := "renew:" + event.RequestID

	var tenantAccountID string
	if tenant.NombaAccountID != nil {
		tenantAccountID = *tenant.NombaAccountID
	}

	_, chargeErr := nombaClient.Charge(ctx, nomba.TokenizedCardChargeRequest{
		Order: nomba.TokenizedCardChargeOrder{
			OrderReference: merchantTxRef,
			CustomerID:     event.CustomerID,
			CallbackURL:    "",
			CustomerEmail:  customer.Email,
			Amount:         fmt.Sprintf("%.2f", float64(plan.Amount)/100.0),
			Currency:       plan.Currency,
			AccountID:      tenantAccountID,
		},
		TokenKey: *customer.NombaTokenKey,
	})

	if chargeErr != nil {
		// Charge failed — mark past_due. The retry queue will redeliver and
		// the scheduler will re-attempt on the next billing sweep.
		log.Error().
			Err(chargeErr).
			Str("subscription_id", event.SubscriptionID).
			Msg("substate renew: card charge failed — marking past_due")
		if serr := setStatus(ctx, q, subID, "past_due", log); serr != nil {
			return fmt.Errorf("substate renew: set past_due: %w (charge error: %w)", serr, chargeErr)
		}
		// Return the charge error so the idempotency wrapper records this as
		// 'failed' and the message is routed to the retry queue.
		return fmt.Errorf("substate renew: charge failed: %w", chargeErr)
	}

	// ── Charge succeeded ──────────────────────────────────────────────────────
	log.Info().
		Str("subscription_id", event.SubscriptionID).
		Str("merchant_tx_ref", merchantTxRef).
		Msg("substate renew: card charge succeeded — advancing period")

	newStart, newEnd := nextPeriod(sub, plan)

	updated, err := q.AdvanceSubscriptionPeriod(ctx, db.AdvanceSubscriptionPeriodParams{
		ID:             subID,
		NewPeriodStart: pgtype.Timestamptz{Time: newStart, Valid: true},
		NewPeriodEnd:   pgtype.Timestamptz{Time: newEnd, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("substate renew: advance period: %w", err)
	}

	log.Info().
		Str("subscription_id", updated.ID.String()).
		Str("status", updated.Status).
		Time("period_start", newStart).
		Time("period_end", newEnd).
		Msg("substate renew: subscription renewed successfully")
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// nextPeriod computes the next billing window for a subscription.
// If the subscription has an existing current_period_end, the new period
// starts from there (so no billing gap). Otherwise it starts from now.
func nextPeriod(sub db.Subscription, plan db.Plan) (start, end time.Time) {
	if sub.CurrentPeriodEnd.Valid && !sub.CurrentPeriodEnd.Time.IsZero() {
		start = sub.CurrentPeriodEnd.Time
	} else {
		start = time.Now().UTC()
	}
	switch plan.Interval {
	case "year":
		end = start.AddDate(1, 0, 0)
	default: // "month"
		end = start.AddDate(0, 1, 0)
	}
	return start, end
}

// setStatus is a thin helper that calls SetSubscriptionStatus and logs the result.
func setStatus(ctx context.Context, q *db.Queries, subID uuid.UUID, status string, log zerolog.Logger) error {
	_, err := q.SetSubscriptionStatus(ctx, db.SetSubscriptionStatusParams{
		ID:     subID,
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("set subscription status %q: %w", status, err)
	}
	log.Info().Str("subscription_id", subID.String()).Str("status", status).Msg("substate: status updated")
	return nil
}
