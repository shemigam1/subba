-- name: CreatePortalToken :one
INSERT INTO portal_access_tokens (tenant_id, customer_id, token_hash, expires_at)
VALUES (sqlc.arg('tenant_id'), sqlc.arg('customer_id'), sqlc.arg('token_hash'), sqlc.arg('expires_at'))
RETURNING *;

-- Resolve + consume a magic-link token. Runs on the admin pool (pre-session).
-- name: GetPortalTokenByHash :one
SELECT * FROM portal_access_tokens WHERE token_hash = $1;

-- name: ConsumePortalToken :execrows
UPDATE portal_access_tokens SET used_at = now()
WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now();
