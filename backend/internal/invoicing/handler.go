// Package invoicing implements the invoicing consumer handler.
//
// The handler is called by the idempotency wrapper with a live Postgres
// transaction (tx). All DB writes go through that tx; the wrapper commits
// or rolls back. The handler never touches the processed_events table or
// does any deduplication itself.
//
// # Event routing
//
//   - payment.succeeded → MarkInvoicePaid (or create-then-pay if no open invoice exists)
//   - subscription.renew → CreateInvoice + CreateInvoiceItem (prorated to the minute)
package invoicing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/shamigam1/subba/internal/broker"
	"github.com/shamigam1/subba/internal/idempotency"
	"github.com/shamigam1/subba/internal/store/db"
)

// NewHandler returns the idempotency.HandlerFunc for the invoicing consumer.
func NewHandler(log zerolog.Logger) idempotency.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, event broker.Event) error {
		q := db.New(tx)
		switch event.EventType {
		case broker.RoutingKeyPaymentSucceeded:
			return handlePaymentSucceeded(ctx, q, event, log)
		case broker.RoutingKeyPaymentReversal:
			return handlePaymentReversal(ctx, q, event, log)
		case broker.RoutingKeyPaymentFailed:
			return handlePaymentFailed(ctx, q, event, log)
		case broker.RoutingKeySubscriptionRenew:
			return handleSubscriptionRenew(ctx, q, event, log)
		default:
			return fmt.Errorf("invoicing: unhandled event type %q", event.EventType)
		}
	}
}

// ── payment.succeeded ─────────────────────────────────────────────────────────

// handlePaymentSucceeded marks the open invoice for this subscription as paid.
// If no open invoice exists (e.g. the renew handler hasn't run yet due to queue
// ordering), we create one inline so the accounting record is always present.
func handlePaymentSucceeded(ctx context.Context, q *db.Queries, event broker.Event, log zerolog.Logger) error {
	subID, err := uuid.Parse(event.SubscriptionID)
	if err != nil {
		return fmt.Errorf("invoicing: parse subscription_id: %w", err)
	}
	tenantID, err := uuid.Parse(event.TenantID)
	if err != nil {
		return fmt.Errorf("invoicing: parse tenant_id: %w", err)
	}
	customerID, err := uuid.Parse(event.CustomerID)
	if err != nil {
		return fmt.Errorf("invoicing: parse customer_id: %w", err)
	}

	nombaRef := &event.NombaReference
	if event.NombaReference == "" {
		nombaRef = nil
	}

	// Try to find an open invoice created by the renew handler.
	subPgID := pgtype.UUID{Bytes: subID, Valid: true}
	inv, err := q.GetOpenInvoiceBySubscription(ctx, subPgID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("invoicing: look up open invoice: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		// No open invoice — create one now so we always have an accounting record.
		// The subscription and plan details are not available here without extra
		// queries (they live in the sub-state handler's domain), so we create a
		// minimal "payment received" invoice using the amount from the event.
		log.Warn().
			Str("subscription_id", event.SubscriptionID).
			Msg("invoicing: no open invoice found for payment.succeeded — creating inline")

		now := time.Now().UTC()
		nowTS := pgtype.Timestamptz{Time: now, Valid: true}
		inv, err = q.CreateInvoice(ctx, db.CreateInvoiceParams{
			TenantID:       tenantID,
			SubscriptionID: pgtype.UUID{Bytes: subID, Valid: true},
			CustomerID:     customerID,
			Amount:         event.Amount,
			Currency:       event.Currency,
			PeriodStart:    nowTS,
			PeriodEnd:      nowTS, // unknown period; sub-state handler will set it
			NombaReference: nombaRef,
		})
		if err != nil {
			return fmt.Errorf("invoicing: create inline invoice: %w", err)
		}

		// Single line item for the inline invoice.
		_, err = q.CreateInvoiceItem(ctx, db.CreateInvoiceItemParams{
			TenantID:    tenantID,
			InvoiceID:   inv.ID,
			Description: "Payment received",
			Amount:      event.Amount,
			Quantity:    1,
			PeriodStart: nowTS,
			PeriodEnd:   nowTS,
		})
		if err != nil {
			return fmt.Errorf("invoicing: create inline invoice item: %w", err)
		}
	}

	// Mark the invoice paid.
	paid, err := q.MarkInvoicePaid(ctx, db.MarkInvoicePaidParams{
		ID:             inv.ID,
		NombaReference: nombaRef,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Invoice was already marked paid (concurrent processing). Not an error —
			// the idempotency wrapper guarantees we only run once, but the invoice
			// may have been paid by an earlier successful run that crashed before ack.
			log.Warn().Str("invoice_id", inv.ID.String()).Msg("invoicing: invoice already paid — skipping")
			return nil
		}
		return fmt.Errorf("invoicing: mark invoice paid: %w", err)
	}

	log.Info().
		Str("invoice_id", paid.ID.String()).
		Str("subscription_id", event.SubscriptionID).
		Str("nomba_reference", event.NombaReference).
		Int64("amount_kobo", paid.Amount).
		Msg("invoicing: invoice marked paid")
	return nil
}

// ── payment.reversal ────────────────────────────────────────────────────────

func handlePaymentReversal(ctx context.Context, q *db.Queries, event broker.Event, log zerolog.Logger) error {
	subID, err := uuid.Parse(event.SubscriptionID)
	if err != nil {
		return fmt.Errorf("invoicing: parse subscription_id for reversal: %w", err)
	}

	if event.NombaReference == "" {
		log.Warn().
			Str("subscription_id", subID.String()).
			Msg("invoicing: payment reversed but no nomba_reference provided")
		return nil
	}

	// Try to find the paid invoice by its nomba reference.
	nombaRef := event.NombaReference
	inv, err := q.GetPaidInvoiceByNombaReference(ctx, &nombaRef)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn().
				Str("subscription_id", subID.String()).
				Str("nomba_reference", event.NombaReference).
				Msg("invoicing: payment reversed but no matching paid invoice found")
			return nil
		}
		return fmt.Errorf("invoicing: get paid invoice by nomba reference: %w", err)
	}

	// Revert the invoice back to 'open'
	_, err = q.MarkInvoiceUnpaid(ctx, inv.ID)
	if err != nil {
		return fmt.Errorf("invoicing: mark invoice unpaid: %w", err)
	}

	log.Info().
		Str("invoice_id", inv.ID.String()).
		Str("subscription_id", subID.String()).
		Str("nomba_reference", event.NombaReference).
		Msg("invoicing: payment reversed — invoice unmarked")
	return nil
}

// ── payment.failed ──────────────────────────────────────────────────────────

func handlePaymentFailed(ctx context.Context, q *db.Queries, event broker.Event, log zerolog.Logger) error {
	log.Info().
		Str("subscription_id", event.SubscriptionID).
		Msg("invoicing: payment failed, invoice remains open")
	return nil
}

// ── subscription.renew ────────────────────────────────────────────────────────

// handleSubscriptionRenew creates the invoice and line items for an upcoming
// billing cycle. The plan amount is taken from event.Amount (set by the
// scheduler) and the billing period is calculated using the plan interval
// stored on the subscription.
//
// Proration to the minute: if the customer's current period doesn't align with
// the calendar month boundary (e.g. they upgraded mid-cycle), we calculate the
// exact prorated fraction as minutes_remaining / minutes_in_period * plan_amount.
func handleSubscriptionRenew(ctx context.Context, q *db.Queries, event broker.Event, log zerolog.Logger) error {
	subID, err := uuid.Parse(event.SubscriptionID)
	if err != nil {
		return fmt.Errorf("invoicing renew: parse subscription_id: %w", err)
	}
	tenantID, err := uuid.Parse(event.TenantID)
	if err != nil {
		return fmt.Errorf("invoicing renew: parse tenant_id: %w", err)
	}
	customerID, err := uuid.Parse(event.CustomerID)
	if err != nil {
		return fmt.Errorf("invoicing renew: parse customer_id: %w", err)
	}

	// Fetch the subscription to get the current period and plan.
	sub, err := q.GetSubscription(ctx, subID)
	if err != nil {
		return fmt.Errorf("invoicing renew: get subscription: %w", err)
	}

	// Fetch the plan for amount and interval.
	plan, err := q.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("invoicing renew: get plan: %w", err)
	}

	// Calculate the next billing period.
	now := time.Now().UTC()
	var periodStart, periodEnd time.Time

	// If the subscription has an existing period, advance from period_end.
	// Otherwise start from now (first billing).
	if sub.CurrentPeriodEnd.Valid && !sub.CurrentPeriodEnd.Time.IsZero() {
		periodStart = sub.CurrentPeriodEnd.Time
	} else {
		periodStart = now
	}

	switch plan.Interval {
	case "month":
		periodEnd = periodStart.AddDate(0, 1, 0)
	case "year":
		periodEnd = periodStart.AddDate(1, 0, 0)
	default:
		return fmt.Errorf("invoicing renew: unknown plan interval %q", plan.Interval)
	}

	// ── Proration calculation (minute-accurate) ───────────────────────────────
	// If the customer is renewing early (e.g. plan upgrade) we prorate:
	//   prorated_amount = plan.Amount * minutes_remaining / total_minutes_in_period
	// For a regular on-schedule renewal, now ≈ periodStart so there is no
	// proration and the full plan amount is charged.
	billedAmount := plan.Amount // default: full period

	if sub.CurrentPeriodStart.Valid && sub.CurrentPeriodEnd.Valid &&
		!sub.CurrentPeriodStart.Time.IsZero() && !sub.CurrentPeriodEnd.Time.IsZero() {
		totalMinutes := sub.CurrentPeriodEnd.Time.Sub(sub.CurrentPeriodStart.Time).Minutes()
		remainingMinutes := sub.CurrentPeriodEnd.Time.Sub(now).Minutes()

		if totalMinutes > 0 && remainingMinutes > 0 && remainingMinutes < totalMinutes {
			// Prorating: only charge for the unused portion of the current period.
			fraction := remainingMinutes / totalMinutes
			billedAmount = int64(math.Round(float64(plan.Amount) * fraction))
			log.Info().
				Float64("fraction", fraction).
				Int64("prorated_kobo", billedAmount).
				Msg("invoicing renew: prorating invoice amount")
		}
	}

	// Override with event.Amount if the scheduler already computed proration.
	// event.Amount == 0 on subscription.renew means "use plan amount".
	if event.Amount > 0 {
		billedAmount = event.Amount
	}

	// ── Create invoice header ─────────────────────────────────────────────────
	inv, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
		TenantID:       tenantID,
		SubscriptionID: pgtype.UUID{Bytes: subID, Valid: true},
		CustomerID:     customerID,
		Amount:         billedAmount,
		Currency:       plan.Currency,
		PeriodStart:    pgtype.Timestamptz{Time: periodStart, Valid: true},
		PeriodEnd:      pgtype.Timestamptz{Time: periodEnd, Valid: true},
		NombaReference: nil, // not yet collected; paid after tokenized-card charge succeeds
	})
	if err != nil {
		return fmt.Errorf("invoicing renew: create invoice: %w", err)
	}

	// ── Create line item ──────────────────────────────────────────────────────
	description := fmt.Sprintf("%s — %s to %s",
		plan.Name,
		periodStart.Format("2 Jan 2006"),
		periodEnd.Format("2 Jan 2006"),
	)

	_, err = q.CreateInvoiceItem(ctx, db.CreateInvoiceItemParams{
		TenantID:    tenantID,
		InvoiceID:   inv.ID,
		Description: description,
		Amount:      billedAmount,
		Quantity:    1,
		PeriodStart: pgtype.Timestamptz{Time: periodStart, Valid: true},
		PeriodEnd:   pgtype.Timestamptz{Time: periodEnd, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("invoicing renew: create invoice item: %w", err)
	}

	log.Info().
		Str("invoice_id", inv.ID.String()).
		Str("subscription_id", event.SubscriptionID).
		Str("plan", plan.Name).
		Str("interval", plan.Interval).
		Int64("amount_kobo", billedAmount).
		Time("period_start", periodStart).
		Time("period_end", periodEnd).
		Msg("invoicing: renewal invoice created")
	return nil
}
