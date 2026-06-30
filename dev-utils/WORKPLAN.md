# Subba — Two-Developer Work Plan

Phase 0 (foundation: schema, RLS, config, logging, platform wiring, health/readiness)
is **done**. This plan splits the remaining work (Phases 1–5 + product surfaces +
deploy + demo) between **two developers** along a single clean seam: **behind the
broker vs. in front of it**.

## The seam

```
        ───────────────  TRACK A: ENGINE  ───────────────
Nomba ─▶ Webhook ─▶ Topic ─▶ invoicing ─┐
        Receiver    Exchange  payouts    ├─▶ Postgres (shared schema)
                              sub-state ─┘
        + Nomba client (OAuth singleton, charges, transfers, virtual accts)
        + idempotency (requestId, consumer) + scheduler/dunning + retry/DLX

        ───────────────  TRACK B: SURFACES  ──────────────
Tenant Dashboard (auth, plans/customers CRUD, API keys, analytics)
Customer Portal  (invoices, subscription/cancel, card tokenization)
Observability    (Prometheus + Grafana, metrics surface)
Evidence         (k6 load + failure injection)  ·  Deploy (Caddy/TLS, MVP URL)  ·  Demo video
```

**Why this seam:** Track A *is* the fault-tolerance thesis the hackathon scores —
it must never be blocked by UI. Track B is broad but highly parallelizable and can
build against mocks until A's endpoints land. They meet at three frozen contracts.

## Recommended assignment

- **Dev A — Engine.** The reliability core + Nomba money rails. Heaviest backend.
- **Dev B — Surfaces.** Both UIs + their APIs, observability, load tests, deploy, demo.

---

## Contracts to freeze on Day 1 (the unblock)

Both tracks stay parallel only if these three are agreed and committed before either
goes heads-down:

1. **REST API contract** (`docs/openapi.yaml`) — every endpoint the dashboard and
   portal call, with request/response shapes. **Dev B drafts, Dev A reviews.** This
   lets B build UIs against mocks immediately.
2. **Event schema** — the JSON published to the topic exchange (routing keys +
   payload: `requestId`, `tenantId`, `eventType`, amount, references). **Dev A owns.**
   Frozen so consumers and the scheduler agree.
3. **Schema additions** — tables B needs that Phase 0 didn't create (dev/customer
   auth, `api_keys`, sessions). **Dev A adds the migrations** (append-only; A owns the
   money schema), B writes the queries. Agree the columns Day 1.

Also agree once: branch-per-feature off `main`, migrations are **append-only and
sequential** (coordinate the next number in Slack), and a shared `internal/metrics`
package both import (A increments counters in handlers, B builds the dashboard).

---

## Track A — Engine (Dev A)

**Phase 1 — the critical seam**
- `internal/nomba`: OAuth **auth singleton** — cache the 60-min `client_credentials`
  token in Redis with single-flight refresh (no stampede, no 401s under load → success
  metric #1). Stub `charge`, `transfer`, `createVirtualAccount` with stable idempotency keys.
- `internal/webhook`: **HMAC-SHA256 verify** (reject before any work → metric #2) →
  publish to topic exchange with **publisher confirms** → `200` only after broker ack.
- `internal/broker`: declare topic exchange, 3 queues + bindings, per-queue **DLX** and
  retry queues with **TTL backoff** (30s / 2m / 10m / 1h); publisher-confirm helper.
- `internal/idempotency`: `(requestId, consumer)` — Redis fast-check + `processed_events`
  durable; a handler wrapper that makes each side effect at-most-once.

**Phase 2 — consumers (`cmd/worker`)**, each a bounded worker pool w/ its own prefetch:
- `invoicing` — immutable invoice + items, **proration to the minute**.
- `payouts` — **revenue split**: transfer the tenant's cut to their Nomba sub-account
  (the moat; most failure-prone → tune retry/DLQ hardest).
- `subscription-state` — advance the billing period, transition status on funding.

**Phase 3 — scheduler (`cmd/scheduler`)** — daily sweep finds due subscriptions and
publishes charge events into the **same** exchange (renewals + dunning, one pipeline).

**Instrumentation A owns:** per-consumer throughput / success / failure / latency,
retry + backoff counts, Nomba call latency/error rate, queue + DLQ depth wiring.

## Track B — Surfaces (Dev B)

**Tenant dashboard** (Phase 4a)
- Backend: dev auth (signup/login + JWT/session), **plans CRUD**, **customers CRUD**,
  **API-key generation** (store only the hash → `tenants.api_key_hash`), analytics endpoints.
- Frontend: dashboard layout + analytics overview, plan/customer management, key display.

**Customer portal** (Phase 4b — Semilore's PRD slice)
- Backend endpoints: customer auth, list invoices, get subscription, **cancel flow**,
  save-card-token callback.
- Frontend: public portal, **invoice-history table**, subscription status + cancel,
  **Nomba card-tokenization** to update an expired card.

**Observability surface** (Phase 5a)
- Prometheus + Grafana in compose; **one Grafana dashboard** (consumer throughput,
  **DLQ depth** as the top fault signal, retry counts, Nomba latency, HTTP p50/95/99,
  PG pool + Redis saturation). HTTP metrics middleware. `/healthz` `/readyz` already done.

**Evidence + deploy** (Phase 5b / 6)
- **k6**: sync profile (ramp API/webhook, p95/p99) + async profile (flood funding
  webhooks, measure end-to-end fanout + backlog drain).
- **Failure injection**: mock payout endpoint failing X% + kill-a-consumer mid-run →
  show the other two keep processing, payouts drain to DLQ, no double-billing → metric #4.
- **Deploy**: Caddy TLS + public HTTPS webhook URL, Dockerfiles for api/worker/scheduler,
  hosted **MVP URL**. **Demo video** (Semilore).

---

## Dependency graph (what blocks what)

```
Day-1 contracts ──┬─▶ A: Phase 1 ──▶ A: Phase 2 ──▶ A: Phase 3
                  │        │              │
                  │        │ (API stubs)  │ (metrics counters)
                  │        ▼              ▼
                  └─▶ B: UIs on mocks ─▶ B: UIs on real API ─▶ B: Grafana ─▶ B: k6 + failure inj ─▶ Deploy + Demo
```

- B is **never hard-blocked**: builds against the frozen contract + mocks until A's
  endpoints exist (~mid-week), then swaps the base URL.
- B's k6 + failure-injection and the Grafana dashboard **depend on A's Phase 2** being
  live — that's the integration point, not a blocker before it.

## Milestones (1-week sprint)

| When | Milestone | Owner |
|---|---|---|
| **Day 1** | 3 contracts frozen + committed; branch/migration rules set | A + B |
| **Day 2 (M1)** | Auth singleton never-401 (metric #1); webhook HMAC + publish-confirm (metric #2); UIs render on mocks | A / B |
| **Day 4 (M2)** | Full cardless-renewal fanout happy path: 3 consumers + idempotency (**the moat**, metric #3); dashboard CRUD + portal invoices/cancel on real API | A / B |
| **Day 5 (M3)** | Per-consumer retry → DLQ + failure injection demonstrable (metric #4); Grafana live; card tokenization working | A / B |
| **Day 6 (M4)** | Scheduler/dunning; deploy to HTTPS MVP URL; k6 figures captured; arch+security note drafted | A + B |
| **Day 7 (M5)** | Demo video recorded; buffer + polish | B (Semilore) |

## Coverage of the four required deliverables

| Deliverable | Owner |
|---|---|
| Public GitHub repo (clean history) | both |
| Working MVP URL (HTTPS) | B (deploy) |
| Demo video (cardless renewal end-to-end) | B (Semilore) |
| Architecture & security note (auth, webhooks, data handling + load/failure figures) | A (engine + figures) + B (load-test numbers) |

## Risks & mitigations

- **A is the critical path.** If it slips, B drops UI polish and pairs on consumers.
- **Nomba sandbox access unknown.** A builds against a Nomba mock from Day 1 so the
  pipeline is testable without live creds; swap to real creds when available.
- **Metrics seam.** Agree the `internal/metrics` counter names Day 1 so A's
  instrumentation and B's Grafana panels line up without rework.
- **Migration collisions.** Append-only + announce the next number; A owns money
  schema, B owns auth/api-key tables.
