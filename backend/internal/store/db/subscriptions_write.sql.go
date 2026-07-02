// NOTE: generated from subscriptions.sql additions. Run `make sqlc` to fully regenerate.
// New methods: AdvanceSubscriptionPeriod, SetSubscriptionStatus.

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ── AdvanceSubscriptionPeriod ─────────────────────────────────────────────────

const advanceSubscriptionPeriod = `-- name: AdvanceSubscriptionPeriod :one
UPDATE subscriptions
SET status               = 'active',
    current_period_start = $1,
    current_period_end   = $2,
    cancel_at_period_end = false
WHERE id = $3
RETURNING id, tenant_id, customer_id, plan_id, status, current_period_start, current_period_end, cancel_at_period_end, canceled_at, created_at, updated_at
`

type AdvanceSubscriptionPeriodParams struct {
	NewPeriodStart time.Time `json:"new_period_start"`
	NewPeriodEnd   time.Time `json:"new_period_end"`
	ID             uuid.UUID `json:"id"`
}

func (q *Queries) AdvanceSubscriptionPeriod(ctx context.Context, arg AdvanceSubscriptionPeriodParams) (Subscription, error) {
	row := q.db.QueryRow(ctx, advanceSubscriptionPeriod,
		arg.NewPeriodStart,
		arg.NewPeriodEnd,
		arg.ID,
	)
	var i Subscription
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.CustomerID,
		&i.PlanID,
		&i.Status,
		&i.CurrentPeriodStart,
		&i.CurrentPeriodEnd,
		&i.CancelAtPeriodEnd,
		&i.CanceledAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

// ── SetSubscriptionStatus ─────────────────────────────────────────────────────

const setSubscriptionStatus = `-- name: SetSubscriptionStatus :one
UPDATE subscriptions
SET status = $1
WHERE id   = $2
RETURNING id, tenant_id, customer_id, plan_id, status, current_period_start, current_period_end, cancel_at_period_end, canceled_at, created_at, updated_at
`

type SetSubscriptionStatusParams struct {
	Status string    `json:"status"`
	ID     uuid.UUID `json:"id"`
}

func (q *Queries) SetSubscriptionStatus(ctx context.Context, arg SetSubscriptionStatusParams) (Subscription, error) {
	row := q.db.QueryRow(ctx, setSubscriptionStatus, arg.Status, arg.ID)
	var i Subscription
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.CustomerID,
		&i.PlanID,
		&i.Status,
		&i.CurrentPeriodStart,
		&i.CurrentPeriodEnd,
		&i.CancelAtPeriodEnd,
		&i.CanceledAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
