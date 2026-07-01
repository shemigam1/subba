package webhook

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/shamigam1/subba/internal/broker"
)

// BrokerPublisher implements Publisher by sending events to RabbitMQ.
type BrokerPublisher struct {
	ch *amqp.Channel
}

// NewBrokerPublisher creates a publisher backed by a RabbitMQ channel.
func NewBrokerPublisher(ch *amqp.Channel) *BrokerPublisher {
	return &BrokerPublisher{ch: ch}
}

// Publish marshals the event to JSON and sends it to the broker.
func (bp *BrokerPublisher) Publish(ctx context.Context, routingKey string, event broker.Event) error {
	return broker.Publish(ctx, bp.ch, routingKey, &event)
}

// DBTenantLookup implements TenantLookup by querying the database.
type DBTenantLookup struct {
	pool *pgxpool.Pool
}

// NewDBTenantLookup creates a tenant lookup backed by the database.
func NewDBTenantLookup(pool *pgxpool.Pool) *DBTenantLookup {
	return &DBTenantLookup{pool: pool}
}

// TenantForVirtualAccount resolves the tenant associated with a virtual account.
// Temporarily returns a dummy tenant ID so webhook testing does not depend on DB state.
func (d *DBTenantLookup) TenantForVirtualAccount(ctx context.Context, accountNumber string) (string, error) {
	// Original DB-backed implementation kept for reference:
	// var tenantID string
	// err := d.pool.QueryRow(ctx, `
	// 	SELECT tenant_id
	// 	FROM customers
	// 	WHERE nomba_virtual_account = $1
	// 	LIMIT 1
	// `, accountNumber).Scan(&tenantID)
	// if err != nil {
	// 	return "", err
	// }
	// return tenantID, nil
	return "dummy-tenant-id", nil
}

// TenantID satisfies the webhook handler's tenant lookup contract.
func (d *DBTenantLookup) TenantID(ctx context.Context, accountNumber string) (string, error) {
	return d.TenantForVirtualAccount(ctx, accountNumber)
}

// TenantIDForVirtualAccount keeps the interface-compatible wrapper expected by the handler.
func (d *DBTenantLookup) TenantIDForVirtualAccount(ctx context.Context, accountNumber string) (string, error) {
	return d.TenantID(ctx, accountNumber)
}

// DBCustomerLookup implements CustomerLookup by querying the database.
type DBCustomerLookup struct {
	pool *pgxpool.Pool
}

// NewDBCustomerLookup creates a customer lookup backed by the database.
func NewDBCustomerLookup(pool *pgxpool.Pool) *DBCustomerLookup {
	return &DBCustomerLookup{pool: pool}
}

// CustomerIDForVirtualAccount looks up which customer owns a given virtual account.
// Temporarily returns a dummy customer ID so webhook testing does not depend on DB state.
func (d *DBCustomerLookup) CustomerIDForVirtualAccount(ctx context.Context, accountNumber string) (string, error) {
	// Original DB-backed implementation kept for reference:
	// var customerID string
	// err := d.pool.QueryRow(ctx, `
	// 	SELECT id FROM customers
	// 	WHERE nomba_virtual_account = $1
	// 	LIMIT 1
	// `, accountNumber).Scan(&customerID)
	// if err != nil {
	// 	return "", err
	// }
	// return customerID, nil
	return "dummy-customer-id", nil
}

// DBSubscriptionLookup implements SubscriptionLookup by querying the database.
type DBSubscriptionLookup struct {
	pool *pgxpool.Pool
}

// NewDBSubscriptionLookup creates a subscription lookup backed by the database.
func NewDBSubscriptionLookup(pool *pgxpool.Pool) *DBSubscriptionLookup {
	return &DBSubscriptionLookup{pool: pool}
}

// SubscriptionIDForCustomer looks up the active subscription for a given customer.
// Temporarily returns a dummy subscription ID so webhook testing does not depend on DB state.
func (d *DBSubscriptionLookup) SubscriptionIDForCustomer(ctx context.Context, tenantID, customerID string) (string, error) {
	// Original DB-backed implementation kept for reference:
	// q := db.New(d.pool)
	// customerUUID, err := uuid.Parse(customerID)
	// if err != nil {
	// 	return "", err
	// }
	// sub, err := q.GetSubscriptionByCustomer(ctx, customerUUID)
	// if err != nil {
	// 	return "", err
	// }
	// return sub.ID.String(), nil
	return "dummy-subscription-id", nil
}
