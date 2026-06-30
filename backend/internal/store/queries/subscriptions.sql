-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE id = $1;

-- name: GetSubscriptionByCustomer :one
SELECT * FROM subscriptions
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateSubscription :one
INSERT INTO subscriptions (tenant_id, customer_id, plan_id, status)
VALUES (sqlc.arg('tenant_id'), sqlc.arg('customer_id'), sqlc.arg('plan_id'), 'incomplete')
RETURNING *;

-- name: CancelSubscription :one
UPDATE subscriptions SET
    cancel_at_period_end = sqlc.arg('at_period_end'),
    status               = CASE WHEN sqlc.arg('at_period_end') THEN status ELSE 'canceled' END,
    canceled_at          = CASE WHEN sqlc.arg('at_period_end') THEN canceled_at ELSE now() END
WHERE id = sqlc.arg('id')
RETURNING *;
