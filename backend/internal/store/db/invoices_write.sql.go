// NOTE: generated from invoices.sql additions. Run `make sqlc` to fully regenerate.
// New methods: CreateInvoice, CreateInvoiceItem, MarkInvoicePaid,
// GetOpenInvoiceBySubscription — appended to the existing invoices.sql.go.

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ── CreateInvoice ─────────────────────────────────────────────────────────────

const createInvoice = `-- name: CreateInvoice :one
INSERT INTO invoices (
    tenant_id, subscription_id, customer_id,
    amount, currency, status,
    period_start, period_end,
    nomba_reference
) VALUES (
    $1, $2, $3,
    $4, $5, 'open',
    $6, $7,
    $8
)
RETURNING id, tenant_id, subscription_id, customer_id, amount, currency, status, period_start, period_end, nomba_reference, issued_at, created_at
`

type CreateInvoiceParams struct {
	TenantID       uuid.UUID    `json:"tenant_id"`
	SubscriptionID pgtype.UUID  `json:"subscription_id"` // nullable FK — use pgtype.UUID{} for NULL
	CustomerID     uuid.UUID    `json:"customer_id"`
	Amount         int64        `json:"amount"`
	Currency       string       `json:"currency"`
	PeriodStart    time.Time    `json:"period_start"`
	PeriodEnd      time.Time    `json:"period_end"`
	NombaReference *string      `json:"nomba_reference"`
}

func (q *Queries) CreateInvoice(ctx context.Context, arg CreateInvoiceParams) (Invoice, error) {
	row := q.db.QueryRow(ctx, createInvoice,
		arg.TenantID,
		arg.SubscriptionID,
		arg.CustomerID,
		arg.Amount,
		arg.Currency,
		arg.PeriodStart,
		arg.PeriodEnd,
		arg.NombaReference,
	)
	var i Invoice
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.SubscriptionID,
		&i.CustomerID,
		&i.Amount,
		&i.Currency,
		&i.Status,
		&i.PeriodStart,
		&i.PeriodEnd,
		&i.NombaReference,
		&i.IssuedAt,
		&i.CreatedAt,
	)
	return i, err
}

// ── CreateInvoiceItem ─────────────────────────────────────────────────────────

const createInvoiceItem = `-- name: CreateInvoiceItem :one
INSERT INTO invoice_items (
    tenant_id, invoice_id, description,
    amount, quantity, period_start, period_end
) VALUES (
    $1, $2, $3,
    $4, $5,
    $6, $7
)
RETURNING id, tenant_id, invoice_id, description, amount, quantity, period_start, period_end, created_at
`

type CreateInvoiceItemParams struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	InvoiceID   uuid.UUID `json:"invoice_id"`
	Description string    `json:"description"`
	Amount      int64     `json:"amount"`
	Quantity    int32     `json:"quantity"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

func (q *Queries) CreateInvoiceItem(ctx context.Context, arg CreateInvoiceItemParams) (InvoiceItem, error) {
	row := q.db.QueryRow(ctx, createInvoiceItem,
		arg.TenantID,
		arg.InvoiceID,
		arg.Description,
		arg.Amount,
		arg.Quantity,
		arg.PeriodStart,
		arg.PeriodEnd,
	)
	var i InvoiceItem
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.InvoiceID,
		&i.Description,
		&i.Amount,
		&i.Quantity,
		&i.PeriodStart,
		&i.PeriodEnd,
		&i.CreatedAt,
	)
	return i, err
}

// ── MarkInvoicePaid ───────────────────────────────────────────────────────────

const markInvoicePaid = `-- name: MarkInvoicePaid :one
UPDATE invoices
SET status          = 'paid',
    nomba_reference = COALESCE($1, nomba_reference)
WHERE id = $2
  AND status = 'open'
RETURNING id, tenant_id, subscription_id, customer_id, amount, currency, status, period_start, period_end, nomba_reference, issued_at, created_at
`

type MarkInvoicePaidParams struct {
	NombaReference *string   `json:"nomba_reference"`
	ID             uuid.UUID `json:"id"`
}

func (q *Queries) MarkInvoicePaid(ctx context.Context, arg MarkInvoicePaidParams) (Invoice, error) {
	row := q.db.QueryRow(ctx, markInvoicePaid, arg.NombaReference, arg.ID)
	var i Invoice
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.SubscriptionID,
		&i.CustomerID,
		&i.Amount,
		&i.Currency,
		&i.Status,
		&i.PeriodStart,
		&i.PeriodEnd,
		&i.NombaReference,
		&i.IssuedAt,
		&i.CreatedAt,
	)
	return i, err
}

// ── GetOpenInvoiceBySubscription ──────────────────────────────────────────────

const getOpenInvoiceBySubscription = `-- name: GetOpenInvoiceBySubscription :one
SELECT id, tenant_id, subscription_id, customer_id, amount, currency, status, period_start, period_end, nomba_reference, issued_at, created_at
FROM invoices
WHERE subscription_id = $1
  AND status = 'open'
ORDER BY issued_at DESC
LIMIT 1
`

func (q *Queries) GetOpenInvoiceBySubscription(ctx context.Context, subscriptionID pgtype.UUID) (Invoice, error) {
	row := q.db.QueryRow(ctx, getOpenInvoiceBySubscription, subscriptionID)
	var i Invoice
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.SubscriptionID,
		&i.CustomerID,
		&i.Amount,
		&i.Currency,
		&i.Status,
		&i.PeriodStart,
		&i.PeriodEnd,
		&i.NombaReference,
		&i.IssuedAt,
		&i.CreatedAt,
	)
	return i, err
}

// pin imports
var _ = pgtype.Timestamptz{}
var _ = time.Time{}
