-- name: ListInvoicesByCustomer :many
SELECT * FROM invoices
WHERE customer_id = $1
ORDER BY issued_at DESC;

-- name: GetInvoice :one
SELECT * FROM invoices WHERE id = $1;

-- name: ListInvoiceItems :many
SELECT * FROM invoice_items WHERE invoice_id = $1 ORDER BY created_at;

-- name: CreateInvoice :one
-- Inserts an immutable invoice header. Status starts as 'open'.
-- amount is in kobo (minor units). period_start/period_end use the subscription
-- billing window so the invoice is human-readable for proration display.
INSERT INTO invoices (
    tenant_id, subscription_id, customer_id,
    amount, currency, status,
    period_start, period_end,
    nomba_reference
) VALUES (
    sqlc.arg('tenant_id'), sqlc.arg('subscription_id'), sqlc.arg('customer_id'),
    sqlc.arg('amount'), sqlc.arg('currency'), 'open',
    sqlc.arg('period_start'), sqlc.arg('period_end'),
    sqlc.arg('nomba_reference')
)
RETURNING *;

-- name: CreateInvoiceItem :one
-- Inserts a single line item linked to an invoice.
-- amount may be negative (proration credit).
INSERT INTO invoice_items (
    tenant_id, invoice_id, description,
    amount, quantity, period_start, period_end
) VALUES (
    sqlc.arg('tenant_id'), sqlc.arg('invoice_id'), sqlc.arg('description'),
    sqlc.arg('amount'), sqlc.arg('quantity'),
    sqlc.arg('period_start'), sqlc.arg('period_end')
)
RETURNING *;

-- name: MarkInvoicePaid :one
-- Transitions an invoice from 'open' → 'paid' and records the Nomba
-- transaction reference. Returns the updated row so callers can confirm.
UPDATE invoices
SET status          = 'paid',
    nomba_reference = COALESCE(sqlc.arg('nomba_reference'), nomba_reference)
WHERE id = sqlc.arg('id')
  AND status = 'open'
RETURNING *;

-- name: GetOpenInvoiceBySubscription :one
-- Looks up the most recent 'open' invoice for a subscription. Used by the
-- invoicing handler to implement create-or-find semantics so a redelivered
-- payment.succeeded event updates the existing invoice rather than creating
-- a duplicate. (Idempotency at the insert level is handled by the wrapper;
-- this query guards against the edge-case where the handler runs twice
-- within the same idempotency window due to a crash-after-commit.)
SELECT * FROM invoices
WHERE subscription_id = sqlc.arg('subscription_id')
  AND status = 'open'
ORDER BY issued_at DESC
LIMIT 1;

-- name: MarkInvoiceUnpaid :one
-- Reverts a paid invoice back to 'open' when a payment reversal occurs.
UPDATE invoices
SET status = 'open'
WHERE id = sqlc.arg('id')
  AND status = 'paid'
RETURNING *;

-- name: GetPaidInvoiceByNombaReference :one
-- Finds a paid invoice by its Nomba transaction reference.
SELECT * FROM invoices
WHERE nomba_reference = sqlc.arg('nomba_reference')
  AND status = 'paid'
LIMIT 1;
