-- +goose Up

-- SaaS developer accounts. Holds the tenant's Nomba credentials. This row IS the
-- tenant; tenant-scoped tables below reference it via tenant_id.
CREATE TABLE tenants (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name                text NOT NULL,
    email               citext UNIQUE NOT NULL,
    nomba_account_id    text,
    nomba_client_id     text,
    nomba_client_secret text,           -- encrypted at the app layer before storage
    api_key_hash        text UNIQUE,     -- hash of the tenant's issued API key
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER tenants_set_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- End-users belonging to a tenant.
CREATE TABLE customers (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email                 citext NOT NULL,
    name                  text,
    nomba_token_key       text,          -- tokenized card for recurring charges
    nomba_virtual_account text,          -- virtual account for cardless renewal (the moat)
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, email)
);
CREATE INDEX idx_customers_tenant ON customers (tenant_id);
CREATE TRIGGER customers_set_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Subscription tiers, e.g. "Pro Plan — NGN 5,000/mo". Amounts are minor units (kobo).
CREATE TABLE plans (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        text NOT NULL,
    amount      bigint NOT NULL CHECK (amount >= 0),       -- minor units (kobo)
    currency    char(3) NOT NULL DEFAULT 'NGN',
    interval    text NOT NULL CHECK (interval IN ('month','year')),
    deleted_at  timestamptz,                                -- soft delete
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_plans_tenant_active ON plans (tenant_id) WHERE deleted_at IS NULL;
CREATE TRIGGER plans_set_updated_at BEFORE UPDATE ON plans
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Links a customer to a plan and tracks the current billing period.
CREATE TABLE subscriptions (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_id          uuid NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    plan_id              uuid NOT NULL REFERENCES plans(id),
    status               text NOT NULL DEFAULT 'incomplete'
                           CHECK (status IN ('incomplete','active','past_due','canceled','unpaid')),
    current_period_start timestamptz,
    current_period_end   timestamptz,
    cancel_at_period_end boolean NOT NULL DEFAULT false,
    canceled_at          timestamptz,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_subscriptions_tenant ON subscriptions (tenant_id);
CREATE INDEX idx_subscriptions_customer ON subscriptions (customer_id);
-- Scheduler sweep finds due subscriptions by period end + status.
CREATE INDEX idx_subscriptions_due ON subscriptions (current_period_end)
    WHERE status IN ('active','past_due');
CREATE TRIGGER subscriptions_set_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Immutable accounting records. Never updated in place except status transitions.
CREATE TABLE invoices (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subscription_id uuid REFERENCES subscriptions(id),
    customer_id     uuid NOT NULL REFERENCES customers(id),
    amount          bigint NOT NULL CHECK (amount >= 0),    -- minor units (kobo)
    currency        char(3) NOT NULL DEFAULT 'NGN',
    status          text NOT NULL DEFAULT 'open'
                      CHECK (status IN ('draft','open','paid','void','uncollectible')),
    period_start    timestamptz,
    period_end      timestamptz,
    nomba_reference text,
    issued_at       timestamptz NOT NULL DEFAULT now(),
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_invoices_tenant ON invoices (tenant_id);
CREATE INDEX idx_invoices_customer ON invoices (customer_id);
CREATE INDEX idx_invoices_subscription ON invoices (subscription_id);

-- Line items; proration credits/charges can be negative.
CREATE TABLE invoice_items (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    invoice_id   uuid NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description  text NOT NULL,
    amount       bigint NOT NULL,                            -- minor units; may be negative
    quantity     int NOT NULL DEFAULT 1 CHECK (quantity > 0),
    period_start timestamptz,
    period_end   timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_invoice_items_invoice ON invoice_items (invoice_id);

-- Durable idempotency record. Keyed on (request_id, consumer) so each of the three
-- fanout consumers deduplicates its OWN processing of an event independently —
-- keying on request_id alone would let the first consumer mark the event done and
-- silently starve the other two.
CREATE TABLE processed_events (
    request_id   text NOT NULL,
    consumer     text NOT NULL,
    status       text NOT NULL DEFAULT 'processing'
                   CHECK (status IN ('processing','done','failed')),
    result       jsonb,
    created_at   timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz,
    PRIMARY KEY (request_id, consumer)
);

-- +goose Down
DROP TABLE IF EXISTS processed_events;
DROP TABLE IF EXISTS invoice_items;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS tenants;
