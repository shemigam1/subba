# Subba — Backend (Nomba Subscriptions Engine)

Event-driven recurring-billing engine on Nomba payment primitives. Go + PostgreSQL
(RLS) + Redis + RabbitMQ.

## Layout

```
cmd/api/              HTTP entrypoint (Phase 0: health/readiness; webhooks + APIs later)
internal/config/      env-based configuration
internal/observability/  structured JSON logging (zerolog) + correlation IDs
internal/platform/    dependency wiring (PG/Redis/RabbitMQ) + readiness checks
internal/store/       DB pool + WithTenant (RLS scoping); sqlc output in store/db
internal/store/queries/  hand-written SQL → typed Go via sqlc
migrations/           goose SQL migrations (schema + RLS policies)
```

## Quickstart

```bash
cp .env.example .env            # adjust if needed (Postgres is on host port 5433)
make up                         # start postgres, redis, rabbitmq
make migrate                    # apply schema + RLS (creates the subba_app role)
make sqlc                       # regenerate typed queries (after editing queries/)
make run                        # start the API on :8080
```

Probes: `curl localhost:8080/healthz` (liveness), `curl localhost:8080/readyz`
(checks Postgres, Redis, RabbitMQ — returns 503 if any is down).

## Multi-tenancy & RLS

Two Postgres connections by design:

- **`DATABASE_URL`** — connects as `subba_app` (non-superuser, RLS enforced). All
  tenant data access goes through `store.WithTenant`, which sets `app.tenant_id`
  for the transaction; RLS policies confine every statement to that tenant.
- **`ADMIN_DATABASE_URL`** — superuser, bypasses RLS. Migrations + tenant-agnostic
  work only (signup, auth, webhook→tenant routing).

RLS is **forced** (`FORCE ROW LEVEL SECURITY`) and default-deny: a query with no
tenant set returns zero rows, and cross-tenant writes are rejected by `WITH CHECK`.

## Notes

- Money is stored as `bigint` minor units (kobo), currency `NGN`.
- `processed_events` is keyed on `(request_id, consumer)` — per-consumer idempotency
  for the fanout; intentionally not under RLS (infrastructure-level, cross-tenant).
- Ports: Postgres `5433`, Redis `6379`, RabbitMQ `5672` / UI `15672` / metrics `15692`.
