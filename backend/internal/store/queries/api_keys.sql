-- name: ListAPIKeys :many
SELECT * FROM api_keys
WHERE revoked_at IS NULL
ORDER BY created_at DESC;

-- name: CreateAPIKey :one
INSERT INTO api_keys (tenant_id, name, key_hash, key_prefix)
VALUES (sqlc.arg('tenant_id'), sqlc.arg('name'), sqlc.arg('key_hash'), sqlc.arg('key_prefix'))
RETURNING *;

-- name: RevokeAPIKey :execrows
UPDATE api_keys SET revoked_at = now()
WHERE id = $1 AND revoked_at IS NULL;

-- Resolve a bearer key to its tenant. Runs on the admin pool (pre-tenant-scope).
-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL;

-- name: TouchAPIKey :exec
UPDATE api_keys SET last_used_at = now() WHERE id = $1;
