-- name: ListPlans :many
SELECT * FROM plans
WHERE (sqlc.arg('include_deleted')::bool OR deleted_at IS NULL)
ORDER BY created_at DESC;

-- name: GetPlan :one
SELECT * FROM plans WHERE id = $1 AND deleted_at IS NULL;

-- name: CreatePlan :one
INSERT INTO plans (tenant_id, name, amount, currency, interval)
VALUES (sqlc.arg('tenant_id'), sqlc.arg('name'), sqlc.arg('amount'), sqlc.arg('currency'), sqlc.arg('interval'))
RETURNING *;

-- name: UpdatePlan :one
UPDATE plans SET
    name     = COALESCE(sqlc.narg('name'), name),
    amount   = COALESCE(sqlc.narg('amount'), amount),
    currency = COALESCE(sqlc.narg('currency'), currency),
    interval = COALESCE(sqlc.narg('interval'), interval)
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePlan :execrows
UPDATE plans SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL;
