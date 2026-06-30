-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS citext;     -- case-insensitive emails
-- +goose StatementEnd

-- Application role used by the API and workers for all tenant-scoped access.
-- It is deliberately NOT a superuser and has no BYPASSRLS, so Row-Level Security
-- policies (added in 00003) are enforced on every query it runs.
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'subba_app') THEN
    CREATE ROLE subba_app LOGIN PASSWORD 'subba_app';
  END IF;
END
$$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO subba_app;
-- Tables are created by the migration role (superuser); auto-grant DML on them
-- to subba_app so it can read/write within the bounds RLS allows.
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO subba_app;

-- Bump updated_at on every UPDATE.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS set_updated_at();
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM subba_app;
REVOKE USAGE ON SCHEMA public FROM subba_app;
DROP ROLE IF EXISTS subba_app;
-- +goose StatementEnd
