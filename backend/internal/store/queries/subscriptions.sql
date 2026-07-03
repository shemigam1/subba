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

-- name: AdvanceSubscriptionPeriod :one
-- Stamps the billing period forward by one interval (month or year) and
-- transitions status to 'active'. Called by the subscription_state handler
-- when a payment.succeeded event confirms a funded renewal.
-- new_period_start and new_period_end are computed in Go from the plan interval.
UPDATE subscriptions
SET status               = 'active',
    current_period_start = sqlc.arg('new_period_start'),
    current_period_end   = sqlc.arg('new_period_end'),
    cancel_at_period_end = false
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: SetSubscriptionStatus :one
-- General-purpose status writer used by the subscription_state handler for
-- past_due and unpaid transitions when a payment fails.
UPDATE subscriptions
SET status = sqlc.arg('status')
WHERE id   = sqlc.arg('id')
RETURNING *;

-- name: ListDueSubscriptions :many
-- Returns all active subscriptions whose current billing period has elapsed.
-- The scheduler sweeps these on each tick and publishes a subscription.renew
-- event for each one so the subscription_state handler can charge the card.
SELECT * FROM subscriptions
WHERE status = 'active'
  AND cancel_at_period_end = false
  AND current_period_end IS NOT NULL
  AND current_period_end <= now();
