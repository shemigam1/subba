-- name: ListCustomers :many
SELECT * FROM customers
WHERE (sqlc.narg('q')::text IS NULL
       OR name ILIKE '%' || sqlc.narg('q') || '%'
       OR email ILIKE '%' || sqlc.narg('q') || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg('lim');

-- name: GetCustomer :one
SELECT * FROM customers WHERE id = $1;

-- name: GetCustomerByEmail :one
SELECT * FROM customers WHERE email = $1;

-- name: CreateCustomer :one
INSERT INTO customers (tenant_id, name, email)
VALUES (sqlc.arg('tenant_id'), sqlc.arg('name'), sqlc.arg('email'))
RETURNING *;

-- name: UpdateCustomer :one
UPDATE customers SET
    name  = COALESCE(sqlc.narg('name'), name),
    email = COALESCE(sqlc.narg('email'), email)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: SetCustomerCardToken :one
UPDATE customers SET nomba_token_key = $2 WHERE id = $1 RETURNING *;
