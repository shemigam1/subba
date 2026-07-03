# Project Handoff: Subba Subscription Engine

## 1. What Was Implemented

### Backend (Go / RabbitMQ / PostgreSQL)
The core infrastructure for the Nomba subscription engine is live:
- HTTP API via Gin, backed by PostgreSQL `pgxpool`.
- Asynchronous RabbitMQ publisher/consumers for webhooks.
- Core routes for Tenants (Dashboard) and Customers (Portal).
- Added PATCH /customers/:id to close the final loop.
- Real integration with Nomba API (`CreateVirtualAccount`, `Transfer`, `Charge`) via the `nomba.Client`.

### Frontend (Next.js App Router)
The frontend has been entirely migrated from raw `useEffect` fetches to a robust `@tanstack/react-query` architecture.
- **Dashboard:** Full parity with the Go backend. We implemented `overview`, `plans`, `customers`, `settings`, and `api-keys`.
- **Portal:** The Customer SDK (`hooks.ts`) provides all hooks required for the payment handshake, including `useVirtualAccount` and `useSaveCard`.
- **Global Auth:** Managed automatically by TanStack Query (`useUser()` hook in `lib/hooks/use-user.ts`).

## 2. Architectural Decisions

1. **MSW vs Live API (The `NEXT_PUBLIC_API_MODE` Flag)**
   - **Decision:** We kept `mockServiceWorker.js` intact but conditionally bypassed it via `.env.local`.
   - **Why:** The primary frontend developer cannot run Docker (thus cannot run the Go backend locally). This approach allows them to keep `NEXT_PUBLIC_API_MODE="mock"` to build UI offline, while the rest of the team sets it to `"live"` to test against `localhost:8080`.
2. **TanStack Query for Auth**
   - **Decision:** We bypassed Context/Redux in favor of React Query's native cache for the `/me` user profile.
   - **Why:** The session is securely held in an `httpOnly` cookie managed by the Go backend. The frontend simply fetches the session data and caches it.

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

## 4. Assumptions & Trade-offs
- **Recharts Dependency:** The `overview` page has placeholders for charts (`[ Recharts LineChart Placeholder ]`). We skipped installing and wiring `recharts` to focus on API completeness.
- **Testing Coverage:** We focused on end-to-end integration rather than unit tests for this hackathon push.

## 4. Remaining Work (If Any)
- **UI Polish:** Replace the chart placeholders on the Overview page with actual graphs.
- **Worker Implementations:** The backend still needs robust retry logic in the RabbitMQ consumers.
- **Scheduler:** The cron job that polls for due subscriptions and triggers billing events needs to be finalized.

## 5. How to Continue Development Safely
1. **Frontend Devs without Docker:** Keep `NEXT_PUBLIC_API_MODE="mock"` in `.env.local` to run against MSW.
2. **Backend/Integration Testers:** Set `NEXT_PUBLIC_API_MODE="live"` in `.env.local`, ensure the Go backend is running on port `8080`, and test the full end-to-end flow.
3. When adding new endpoints, always update `v1.d.ts` (using `openapi-typescript` against the Go swagger spec) before creating new `useQuery` or `useMutation` hooks.
