# Subba — Security Model

The goal is a billing engine **Nomba could adopt as a plug-and-play product** — which
means security is a feature, not an afterthought. This is also the basis of the
required *Architecture & Security Note*. Items marked ✅ are implemented in Phase 0;
the rest are committed design for Phases 1–5.

## 1. Multi-tenant isolation (Postgres RLS) ✅

- Every tenant-scoped table has **`FORCE ROW LEVEL SECURITY`** with a default-deny
  policy on `tenant_id = current_tenant_id()`.
- **Two DB roles:** the app connects as the non-superuser `subba_app` (RLS enforced);
  a separate privileged role is used only for migrations and tenant-agnostic work.
- The active tenant is set per transaction via `set_config('app.tenant_id', …, true)`
  in `store.WithTenant`; clients **never** pass a tenant id.
- **Proven**, not asserted: tenant A cannot see tenant B's rows, an unset context
  returns zero rows, and cross-tenant writes are rejected by `WITH CHECK`.

## 2. Secrets handling

- **Tenant Nomba keys** are encrypted at rest at the app layer (AES-GCM with a master
  key from the deploy secret store), **never logged**, and **never returned** by the
  API — `GET /settings` exposes only `*_set` booleans.
- **Tenant API keys** are stored as a hash (`tenants.api_key_hash`); the full secret is
  shown **once** at creation (`ApiKeyCreated.secret`) and is unrecoverable after.
- **No secrets in source control:** `.env` is gitignored; live credentials live only in
  the deploy platform's secret store. (Any key shared over a non-private channel is
  rotated.)
- **No secrets in logs or emails.** Structured logs carry a `request_id`, never tokens
  or card data.

## 3. Webhook authenticity (the money ingress)

- Nomba webhooks: **HMAC-SHA256** signature verified **before any work**; invalid
  signatures rejected at the seam. Only then is the event published (with publisher
  confirms) and 200 returned.
- Resend delivery/bounce webhooks (stretch) are likewise signature-verified.

## 4. Exactly-once money effects

- RabbitMQ delivers at-least-once; idempotency keyed on **`(requestId, consumer)`**
  makes each side effect at-most-once → **exactly-once effect**.
- Charges and transfers to Nomba carry a **stable idempotency key**, so a retry never
  moves money twice. Durable record in `processed_events`.

## 5. Authentication & sessions

- **Tenant Dashboard:** email + password (hashed with bcrypt/argon2id) → JWT/session in
  an **httpOnly, Secure, SameSite=Lax** cookie set by the BFF; token never in client JS.
- **Customer Portal (passwordless):** magic-link tokens —
  - 32-byte CSPRNG, **opaque** (not JWT) so they are revocable and single-use;
  - only **`sha256(token)`** stored (`portal_access_tokens`);
  - **15-minute** expiry, **single-use** (`used_at` enforced);
  - exchanged for a short-lived server-side (Redis) **`subba_portal`** session cookie;
  - scoped to exactly one `customer_id`; all portal queries are RLS-bound to it.
- **Enumeration-safe:** `POST /portal/access-request` always returns `202` whether or
  not the email exists. Rate-limited per email + per IP.
- **Machine-to-machine:** tenant API keys via `Authorization: Bearer`.

## 6. Transactional email (Resend)

- Sending domain verified with **SPF + DKIM + DMARC** (deliverability = the link
  actually arrives, which is itself a trust/security property).
- White-labelled From (`"{Tenant} via Subba"`, tenant `replyTo`); custom per-tenant
  domains are a documented stretch.
- Emails contain only a one-time **HTTPS** link — no tokens-at-rest, no PII beyond the
  recipient, idempotent sends so dunning never double-fires.

## 7. Card data / PCI scope minimization

- Cards are tokenized by **Nomba's hosted tokenization**; raw PAN **never touches
  Subba**. We store only the Nomba token reference (`customers.nomba_token_key`) and
  display metadata (brand/last4/expiry).
- The cardless path avoids cards entirely: funds arrive by bank transfer to a Nomba
  **virtual account**.

## 8. Transport, input & platform

- **TLS everywhere** (Caddy termination), **HSTS**; cookies `Secure`.
- Input validation (zod on the edge, server-side validation on every handler);
  parameterized queries via sqlc (no string-built SQL).
- **CORS** locked to the known dashboard/portal origins.
- **Rate limiting** on auth, access-request, and webhook endpoints.
- **Least privilege**: the app's DB role has only DML within RLS bounds; no DDL/superuser.

## 9. Observability for security

- Correlation `request_id` propagated from webhook receipt through all consumers, so a
  single payment is traceable end to end without leaking secrets.
- `/readyz` reports per-dependency health (graceful degradation, demonstrable live).
- Dashboard surfaces **DLQ depth** — a rising dead-letter queue is the primary signal
  that something in the money path is failing.

## Schema additions this implies (for the Day-1 migration)

| Table | Purpose |
|---|---|
| `tenant_auth` (or columns on `tenants`) | `password_hash` for dashboard login |
| `api_keys` | `id, tenant_id, name, key_hash, created_at, last_used_at, revoked_at` |
| `portal_access_tokens` | `customer_id, token_hash, expires_at, used_at, created_at` |
| `sessions` (or Redis-only) | dashboard/portal session store |

Track A owns these migrations (append-only); columns above are the agreed shape.
