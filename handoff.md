# Project Handoff: Subba Subscription Engine

## 1. What Was Implemented

### Backend (Go / RabbitMQ / PostgreSQL)
The core infrastructure for the Nomba subscription engine is live:
- HTTP API via Gin, backed by PostgreSQL `pgxpool`.
- Asynchronous RabbitMQ publisher/consumers for webhooks.
- Core routes for Tenants (Dashboard) and Customers (Portal).
- Added PATCH /customers/:id to close the final loop.
- **Virtual Account Provisioning:** Customers synchronously receive Nomba Virtual Accounts (tagged with custom `{tenantID}:{customerID}` `accountRef` markers). *Recently updated to strictly match the Nomba API payload (`currency`, `expectedAmount`) and correct path (`/v1/accounts/virtual/{subAccountId}`), as well as correctly parse the returned `bankAccountNumber`.*
- **O(1) Webhook Processing:** Webhook handlers immediately parse the incoming `accountRef` to bypass DB lookups completely.
- **Webhook Signature Verification:** Nomba signs the raw HTTP body. Subba performs standard HMAC-SHA256 and HMAC-SHA512 validation of the raw payload against the `nomba-signature` header before proceeding to parse the JSON. This ensures payload integrity and guards against spoofing.
- **Checkout & Tokenized Cards:** Fully implemented `/v1/checkout/order` to generate checkout links and tokenize cards, and updated our recurring billing trigger to correctly use `/v1/checkout/tokenized-card-payment`.
- **Scheduler Process:** A cron-driven sweep service running in Go that publishes renewal events to RabbitMQ.

### Frontend (Next.js App Router)
The frontend has been entirely migrated from raw `useEffect` fetches to a robust `@tanstack/react-query` architecture.
- **Dashboard:** Full parity with the Go backend. We implemented `overview`, `plans`, `customers`, `settings`, and `api-keys`.
- **Portal:** The Customer SDK (`hooks.ts`) provides all hooks required for the payment handshake, including `useVirtualAccount` and `useSaveCard`.
- **Global Auth:** Managed automatically by TanStack Query (`useUser()` hook in `lib/hooks/use-user.ts`).

## 2. Architectural Decisions

1. **MSW vs Live API (The `NEXT_PUBLIC_API_MODE` Flag)**
   - **Decision:** We kept `mockServiceWorker.js` intact but conditionally bypassed it via `.env.local`.
   - **Why:** The primary frontend developer cannot run Docker (thus cannot run the Go backend locally). This approach allows them to keep `NEXT_PUBLIC_API_MODE="mock"` to build UI offline, while the rest of the team sets it to `"live"` to test against `localhost:8080`.
2. **TanStack Query for Auth & Bearer Tokens**
   - **Decision:** We use React Query's native cache for the `/me` user profile, and session IDs are passed via `Authorization: Bearer <token>`.
   - **Why:** Because the frontend (Vercel) and API (CloudFront) are cross-site, `httpOnly` cookies get dropped by the browser. We updated the backend to return session tokens in the login/signup JSON payloads and updated the Go middlewares to seamlessly parse Bearer session tokens alongside legacy API keys.
3. **Instant Settlement over Manual Payouts**
   - **Decision:** The backend does NOT use a Payouts consumer and does NOT manually initiate `Transfers` to pay out funds to tenants. The obsolete Payouts RabbitMQ topology was permanently torn down.
   - **Why:** Because we scope all `VirtualAccount` creations to the tenant's `subAccountID`, Nomba's Instant Settlement automatically credits the tenant's balance instantly. Any attempt to manually implement a Payouts Worker would result in double-paying the tenant.

## 3. UI Map for Testing Backend Features

Backend engineers should use the following frontend locations to test their API changes end-to-end:

### Merchant Dashboard (`http://localhost:3000`)
- **Login / Signup:** Test `POST /auth/login` and `POST /auth/signup` at `/login` and `/signup`.
- **Plans (CRUD):** Navigate to **Plans** (`/plans`) to test `GET /plans`, `POST /plans` (New Plan drawer), `PATCH /plans/:id` (Edit pencil icon), and `DELETE /plans/:id` (Trash icon).
- **Customers (CRUD):** Navigate to **Customers** (`/customers`) to test `GET /customers` and `POST /customers` (Add Customer drawer).
- **Customer Details:** Click "View" on any customer to go to `/customers/[id]`. Here you can test:
  - `PATCH /customers/:id` ("Edit Profile" button)
  - `POST /customers/:id/portal-link` ("Generate Portal Link" button)
  - `POST /subscriptions` and `POST /subscriptions/:id/cancel` (Subscription Active/Create section)
  - `GET /customers/:id/invoices` (Invoices table at the bottom)
- **API Keys:** Navigate to **Developers > API Keys** (`/api-keys`) to test generating (`POST /api-keys`) and revoking (`DELETE /api-keys/:id`).
- **Settings:** Navigate to **Settings** (`/settings`) to test patching webhook endpoints (`PATCH /settings`).

### Customer Portal
To test the customer portal, go to a customer's detail page in the dashboard and click **Generate Portal Link**. Open that link in a new incognito window (or different browser) to isolate cookies.
- **Home (`/pay`):** Tests `GET /portal/subscription` and `POST /portal/subscription/cancel`.
- **Payment Method (`/pay/payment-method`):** Tests `GET /portal/virtual-account` and saving a tokenized card (`POST /portal/payment-method/card`).
- **Invoices (`/pay/invoices`):** Tests `GET /portal/invoices`.
- **Checkout:** Implemented — the invoice detail page (`/pay/invoices/[id]`) shows a **"Pay with card"** button on any non-paid invoice. It calls `POST /portal/invoices/{id}/checkout` and redirects the browser to the returned Nomba `checkoutLink`. The endpoint is documented in `dev-utils/openapi.yaml` and typed in `lib/api/v1.d.ts`.

## 4. Assumptions & Trade-offs
- **Recharts Dependency:** The `overview` page has placeholders for charts (`[ Recharts LineChart Placeholder ]`). We skipped installing and wiring `recharts` to focus on API completeness.
- **Testing Coverage:** We focused on end-to-end integration rather than unit tests for this hackathon push.

## 5. Remaining Work (If Any)
- **UI Polish:** Replace the chart placeholders on the Overview page with actual graphs.
- **Checkout reconciliation:** after a hosted-checkout payment, the invoice flips to `paid` only via the Nomba webhook → invoicing consumer path. The portal invoice page doesn't poll; the customer sees the update on refresh.
- **`next.config.ts` suppresses TS/ESLint errors during builds** (`ignoreBuildErrors`, committed for Vercel). Compensating control: keep `npx tsc --noEmit` clean locally — it is, except one known benign error in `next.config.ts` itself (`eslint` key no longer in Next 16's `NextConfig` type).
- **Local dev papercut:** `goose@latest` now requires Go ≥ 1.25.7, while local dev pins `GOTOOLCHAIN=go1.25.0` (to match CI). Workaround: `GOTOOLCHAIN=auto make migrate`. Proper fix: pin a goose version in the Makefile.

## 5b. Audit log (2026-07-06)
A full backend-vs-contract-vs-frontend audit found and fixed:
1. **[critical] Dashboard Bearer auth was broken** — `RequireTenant` gated the session lookup on `uuid.Parse(token)`, but session ids are 43-char base64url strings (`auth.RandomToken`), never UUIDs, so every dashboard Bearer call 401'd (login → instant redirect loop in the cross-site deployment). Fixed: session lookup first (Redis), fall through to API-key hash. Verified live: session Bearer 200, API-key Bearer 200, garbage 401, portal Bearer 200.
2. **Float math on money** in `CreateCheckoutLink` (`fmt.Sprintf("%.2f", float64(amount)/100)`) → replaced with integer math (`%d.%02d`).
3. **Checkout endpoint was undocumented and unused** — added to `dev-utils/openapi.yaml`, regenerated `v1.d.ts`, wired a "Pay with card" button on the portal invoice detail page.
4. **No dashboard sign-out existed** — sidebar footer showed a hardcoded "Jane Doe"; now shows the real tenant (from `useUser`) with a working sign-out (POST /auth/logout + clear token + redirect).
5. **Phantom fields on the customer detail page** — `phone` (collected/PATCHed but not in the schema; silently dropped), invoice `created_at` (API sends `issued_at`; the Date column rendered garbage), `invoice_url` (never returned; dead column). All removed; every call on that page is now typed (no `api as any` remains in the app).
Verified after fixes: `go build`/`go vet` clean, full backend test suite green against live infra, `tsc` clean (except the known next.config quirk), `next build` passes (15 routes).

## 6. How to Continue Development Safely
1. **Nomba Credentials Safety:** The backend `config.go` was previously configured to crash if live credentials were used, but this safety check has been removed as per Hackathon guidelines to allow live testing on `api.nomba.com`. Ensure you only use test accounts when interacting with the live system during development.
2. **Simulating Webhooks:** You do not need the live Nomba system to test webhook processing. Simply run `node dev-utils/test_webhook.js` to simulate properly signed payloads against your local backend.
3. **Frontend Devs without Docker:** Keep `NEXT_PUBLIC_API_MODE="mock"` in `.env.local` to run against MSW.
4. **Backend/Integration Testers:** Set `NEXT_PUBLIC_API_MODE="live"` in `.env.local`, ensure the Go backend is running on port `8080`, and test the full end-to-end flow.
5. When adding new endpoints, always update `v1.d.ts` (using `openapi-typescript` against the Go swagger spec) before creating new `useQuery` or `useMutation` hooks.
