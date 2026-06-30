-- Starter queries to validate the sqlc pipeline. Real query sets per domain
-- (customers, plans, subscriptions, invoices) are added in Phase 1.

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantByAPIKeyHash :one
SELECT * FROM tenants WHERE api_key_hash = $1;

-- name: CreateTenant :one
INSERT INTO tenants (name, email, nomba_account_id, nomba_client_id, nomba_client_secret, api_key_hash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
