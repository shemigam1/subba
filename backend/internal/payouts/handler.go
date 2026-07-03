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

// NewHandler returns the idempotency.HandlerFunc for the payouts consumer.
// nombaClient is the shared Nomba API client; it is safe for concurrent use.
func NewHandler(nombaClient *nomba.Client, log zerolog.Logger) idempotency.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, event broker.Event) error {
		if event.EventType != broker.RoutingKeyPaymentSucceeded {
			// payouts only cares about confirmed inbound payments.
			return fmt.Errorf("payouts: unexpected event type %q", event.EventType)
		}
		return handle(ctx, tx, nombaClient, event, log)
	}
}

func handle(ctx context.Context, tx pgx.Tx, nombaClient *nomba.Client, event broker.Event, log zerolog.Logger) error {
	q := db.New(tx)

	tenantID, err := uuid.Parse(event.TenantID)
	if err != nil {
		return fmt.Errorf("payouts: parse tenant_id: %w", err)
	}
	customerID, err := uuid.Parse(event.CustomerID)
	if err != nil {
		return fmt.Errorf("payouts: parse customer_id: %w", err)
	}

	// ── 1. Fetch tenant to get their sub-account details ─────────────────────
	tenant, err := q.GetTenantByID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("payouts: get tenant: %w", err)
	}

	if tenant.NombaSubaccountID == nil || *tenant.NombaSubaccountID == "" {
		// Tenant has not configured a sub-account yet. Log and skip — this is
		// a business configuration issue, not a transient error, so we treat it
		// as success to avoid infinite retry loops. The invoice record is still
		// created so it shows up in reconciliation.
		log.Warn().
			Str("tenant_id", event.TenantID).
			Msg("payouts: tenant has no nomba_subaccount_id — skipping payout")
		_, err := createPendingPayoutRecord(ctx, q, tenantID, customerID, event)
		return err
	}

	// ── 2. Write the pending payout record (pre-call) ─────────────────────────
	// This creates an open invoice that records intent to pay out. If the
	// handler crashes between here and step 3, the retry will hit the same
	// idempotency slot and re-attempt the transfer call. Nomba's merchantTxRef
	// deduplication ensures no double-payout.
	inv, err := createPendingPayoutRecord(ctx, q, tenantID, customerID, event)
	if err != nil {
		return fmt.Errorf("payouts: create pending record: %w", err)
	}

	// ── 3. Call Nomba Transfer ─────────────────────────────────────────────────
	// merchantTxRef = "payout:{requestID}" — unique per event, safe to retry.
	merchantTxRef := "payout:" + event.RequestID

	transferResp, transferErr := nombaClient.Transfer(ctx, nomba.BankTransferRequest{
		AccountID:     *tenant.NombaSubaccountID,
		Amount:        event.Amount,
		BankCode:      "", // Sub-account routing — Nomba resolves from AccountID
		AccountNumber: "",
		AccountName:   tenant.Name,
		SenderName:    "Subba Platform",
		Narration:     fmt.Sprintf("Subscription revenue — ref %s", event.NombaReference),
		MerchantTxRef: merchantTxRef,
	})

	// ── 4. Update the payout record based on Nomba response ───────────────────
	if transferErr != nil {
		// Non-fatal: leave the invoice 'open' so the retry loop can re-attempt.
		// The wrapper's MarkFailed will stamp processed_events.status='failed'.
		log.Error().
			Err(transferErr).
			Str("request_id", event.RequestID).
			Str("merchant_tx_ref", merchantTxRef).
			Msg("payouts: nomba transfer call failed")
		return fmt.Errorf("payouts: nomba transfer: %w", transferErr)
	}

	// Transfer succeeded — mark the payout invoice paid.
	nombaRef := transferResp.Data.TransactionID
	_, err = q.MarkInvoicePaid(ctx, db.MarkInvoicePaidParams{
		ID:             inv.ID,
		NombaReference: &nombaRef,
	})
	if err != nil {
		return fmt.Errorf("payouts: mark payout invoice paid: %w", err)
	}

	log.Info().
		Str("invoice_id", inv.ID.String()).
		Str("request_id", event.RequestID).
		Str("nomba_transfer_id", nombaRef).
		Str("tenant_id", event.TenantID).
		Int64("amount_kobo", event.Amount).
		Msg("payouts: payout disbursed successfully")
	return nil
}

// createPendingPayoutRecord writes an 'open' invoice that represents a pending
// payout. Returns the invoice so the caller can later mark it paid or leave it
// open for retry.
func createPendingPayoutRecord(
	ctx context.Context,
	q *db.Queries,
	tenantID, customerID uuid.UUID,
	event broker.Event,
) (db.Invoice, error) {
	now := nowUTC()
	description := fmt.Sprintf("Payout to tenant — subscription %s", event.SubscriptionID)

	inv, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
		TenantID:       tenantID,
		SubscriptionID: pgtype.UUID{}, // payout invoices are not tied to a subscription
		CustomerID:     customerID,
		Amount:         event.Amount,
		Currency:       event.Currency,
		PeriodStart:    now,
		PeriodEnd:      now,
		NombaReference: nil,
	})
	if err != nil {
		return db.Invoice{}, fmt.Errorf("create payout invoice: %w", err)
	}

	_, err = q.CreateInvoiceItem(ctx, db.CreateInvoiceItemParams{
		TenantID:    tenantID,
		InvoiceID:   inv.ID,
		Description: description,
		Amount:      event.Amount,
		Quantity:    1,
		PeriodStart: now,
		PeriodEnd:   now,
	})
	if err != nil {
		return db.Invoice{}, fmt.Errorf("create payout invoice item: %w", err)
	}

	return inv, nil
}

// nowUTC returns the current UTC time — extracted for testability.
func nowUTC() time.Time { return time.Now().UTC() }
