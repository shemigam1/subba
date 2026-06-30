-- +goose Up

-- Dashboard login: password hash on the tenant (argon2id at the app layer).
ALTER TABLE tenants ADD COLUMN password_hash text;
-- Settings fields surfaced via /settings (secrets stored encrypted at the app layer).
ALTER TABLE tenants ADD COLUMN nomba_subaccount_id text;
ALTER TABLE tenants ADD COLUMN webhook_url text;
ALTER TABLE tenants ADD COLUMN webhook_secret text;

-- Tenant API keys for server-to-server access. Only the hash is stored; the full
-- secret is shown once at creation.
CREATE TABLE api_keys (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name         text,
    key_hash     text NOT NULL UNIQUE,    -- sha256(secret)
    key_prefix   text NOT NULL,           -- display, e.g. "sk_test_ab12"
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz,
    revoked_at   timestamptz
);
CREATE INDEX idx_api_keys_tenant ON api_keys (tenant_id) WHERE revoked_at IS NULL;

ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON api_keys
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- Passwordless customer portal access tokens (magic links). Single-use, expiring;
-- only the hash is stored.
CREATE TABLE portal_access_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_id uuid NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    token_hash  text NOT NULL UNIQUE,     -- sha256(opaque token)
    expires_at  timestamptz NOT NULL,
    used_at     timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_portal_tokens_customer ON portal_access_tokens (customer_id);

ALTER TABLE portal_access_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE portal_access_tokens FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON portal_access_tokens
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- +goose Down
DROP TABLE IF EXISTS portal_access_tokens;
DROP TABLE IF EXISTS api_keys;
ALTER TABLE tenants DROP COLUMN IF EXISTS webhook_secret;
ALTER TABLE tenants DROP COLUMN IF EXISTS webhook_url;
ALTER TABLE tenants DROP COLUMN IF EXISTS nomba_subaccount_id;
ALTER TABLE tenants DROP COLUMN IF EXISTS password_hash;
