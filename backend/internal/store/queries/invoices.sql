-- name: ListInvoicesByCustomer :many
SELECT * FROM invoices
WHERE customer_id = $1
ORDER BY issued_at DESC;

-- name: GetInvoice :one
SELECT * FROM invoices WHERE id = $1;

-- name: ListInvoiceItems :many
SELECT * FROM invoice_items WHERE invoice_id = $1 ORDER BY created_at;
