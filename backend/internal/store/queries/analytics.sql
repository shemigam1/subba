-- name: CountActiveSubscriptions :one
SELECT count(*) FROM subscriptions WHERE status = 'active';

-- Monthly recurring revenue: normalize yearly plans to a monthly figure (minor units).
-- name: SumMRR :one
SELECT COALESCE(SUM(
    CASE WHEN p.interval = 'year' THEN p.amount / 12 ELSE p.amount END
), 0)::bigint AS mrr
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.status = 'active';

-- name: CountPaymentsToday :one
SELECT count(*) FROM invoices
WHERE status = 'paid' AND issued_at >= date_trunc('day', now());

-- name: CountFailedInvoices :one
SELECT count(*) FROM invoices WHERE status = 'uncollectible';

-- name: RevenueSeries :many
SELECT date_trunc('day', issued_at)::date AS day, COALESCE(SUM(amount), 0)::bigint AS amount
FROM invoices
WHERE status = 'paid' AND issued_at >= now() - interval '30 days'
GROUP BY day
ORDER BY day;
