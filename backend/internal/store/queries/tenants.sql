-- Auth lookups run on the admin pool (pre-session, RLS-bypassing); everything else
-- runs tenant-scoped via WithTenant.

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantByEmail :one
SELECT * FROM tenants WHERE email = $1;

-- name: CreateTenant :one
INSERT INTO tenants (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateTenantSettings :one
UPDATE tenants SET
    name                = COALESCE(sqlc.narg('business_name'), name),
    support_email       = COALESCE(sqlc.narg('support_email'), support_email),
    webhook_url         = COALESCE(sqlc.narg('webhook_url'), webhook_url),
    nomba_account_id    = COALESCE(sqlc.narg('nomba_account_id'), nomba_account_id),
    nomba_subaccount_id = COALESCE(sqlc.narg('nomba_subaccount_id'), nomba_subaccount_id),
    nomba_client_id     = COALESCE(sqlc.narg('nomba_client_id'), nomba_client_id),
    nomba_client_secret = COALESCE(sqlc.narg('nomba_client_secret'), nomba_client_secret)
WHERE id = $1
RETURNING *;

