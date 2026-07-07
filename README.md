# Subba: Nomba Subscriptions Engine

[![Go Report Card](https://goreportcard.com/badge/github.com/shamigam1/subba)](https://goreportcard.com/report/github.com/shamigam1/subba)
[![Next.js](https://img.shields.io/badge/Next.js-14-black)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.22-blue)](https://golang.org/)

An event-driven, fault-tolerant recurring billing layer built natively on Nomba's payment primitives. Subba allows SaaS businesses to abstract away webhook idempotency, proration, and subscription state management, specifically catering to the African market.

## Table of Contents
- [Overview](#overview)
- [Key Features](#key-features)
- [Technology Stack](#technology-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Authentication & Authorization](#authentication--authorization)
- [Setup & Installation](#setup--installation)
- [Running Locally](#running-locally)
- [Database & Migrations](#database--migrations)
- [API Overview](#api-overview)
- [Testing & Observability](#testing--observability)
- [Design Decisions & Trade-offs](#design-decisions--trade-offs)
- [Future Improvements](#future-improvements)
- [License](#license)

## Overview

Implementing recurring billing at scale is notoriously difficult. Developers are forced to build complex math from scratch, manage database connection limits during webhook bursts, and handle failed payment retries (dunning). Furthermore, generic billing engines alienate African customers who lack debit cards, and they do not provide native, automated revenue splitting for marketplace platforms.

**Subba** provides a plug-and-play, enterprise-grade subscriptions engine. We lean heavily into Nomba's African-first primitives: creating a "Cardless Subscription Moat" (using Nomba Virtual Accounts for auto-renewing subscriptions via bank transfers) and handling tokenized card infrastructure seamlessly.

## Live Deployments

- **Live Frontend (Vercel):** https://subba-theta.vercel.app
- **Live API Endpoint (Cloudflare Tunnel):** https://asbestos-dale-serve-checks.trycloudflare.com/v1

## Key Features

- **Event-Driven Core:** A RabbitMQ Topic Exchange fanout topology that processes Nomba webhooks asynchronously, ensuring that invoicing and state management are strictly decoupled.
- **Cardless Subscriptions:** Allows African users to maintain recurring SaaS subscriptions using dynamically provisioned Nomba Virtual Accounts (Direct Bank Transfers).
- **Native Revenue Splitting & Instant Settlement:** Instead of building a brittle, custom payouts engine, Subba relies entirely on Nomba's native Sub-Account architecture. When a subscriber pays, Nomba routes the tenant's cut directly to their sub-account and instantly settles it, completely eliminating double-payout risks.
- **O(1) Webhook Resolution:** When provisioning Virtual Accounts, Subba injects a composite `accountRef` (`{tenantID}:{customerID}`). When Nomba webhooks arrive, the system instantly routes the payment to the correct subscription state without a single database lookup.
- **Fault-Tolerant Consumers:** Granular Dead-Letter Exchanges (DLX) and TTL-based retry policies per consumer prevent head-of-line blocking.
- **Strict Idempotency:** Exactly-once processing logic utilizing Redis fast-paths and Postgres ACID constraints via composite `(requestId, consumer)` keys.
- **Dual-Session Multi-Tenancy:** Hardened logical separation preventing session collisions between the Merchant Dashboard and the Customer Portal in the same browser.
- **Production-Ready Frontend:** An immersive, highly-responsive Next.js frontend built entirely on `@tanstack/react-query` to ensure instant data availability and optimistic UI updates.

## Technology Stack

- **Backend:** Go (Golang) with Gin framework. Chosen for its concurrency model (goroutines), strict typing, and high throughput—perfect for handling webhook storms.
- **Frontend:** Next.js (App Router), React, Tailwind CSS, TanStack Query. Chosen for developer velocity, robust caching, and server-side rendering where applicable.
- **Database:** PostgreSQL (via `pgxpool`). Chosen for ACID compliance and robust Row-Level Security (RLS) features, which act as the absolute source of truth.
- **Caching & Ephemeral State:** Redis. Chosen for blazing fast idempotency locks, rate limiting, and centralized session storage.
- **Message Broker:** RabbitMQ. Chosen for its mature, durable queues and advanced routing (Topic Exchanges, Dead Lettering) to support our event-driven fanout topology.

## Architecture

Subba is split into two primary components: the **Frontend SPA** and the **Go API/Worker Engine**.

1. **The API Edge:** Receives webhooks from Nomba and authenticates incoming Merchant/Customer requests.
2. **The Message Bus (RabbitMQ):** When a webhook arrives (e.g. a transfer succeeds), the API immediately publishes it to an exchange and returns `200 OK` to Nomba.
3. **The Worker Pool:** Independent Go consumers subscribe to the broker. One consumer updates the Subscription State, and another triggers Invoicing and receipts. If processing fails, events hit a DLX and retry later, while other subsystems remain unaffected. (Note: Payouts are handled natively and instantly by Nomba's Sub-Account architecture, abstracting that complexity out of our backend entirely).
4. **The Scheduler:** A standalone cron-driven Go process that continuously sweeps Postgres for active subscriptions whose billing periods have elapsed, publishing `subscription.renew` events into the broker.

## Project Structure

```text
subba/
├── backend/                  # The Go API, Worker Engine, and Scheduler
│   ├── cmd/
│   │   ├── api/              # Entry point for the HTTP server
│   │   ├── scheduler/        # Entry point for the Renewal Scheduler
│   │   └── worker/           # Entry point for the RabbitMQ consumers
│   ├── internal/
│   │   ├── auth/             # Session management (Redis) and password hashing
│   │   ├── http/             # Gin router, middleware, and route handlers
│   │   ├── store/            # PostgreSQL repository layer (pgx)
│   │   ├── webhook/          # RabbitMQ publishers, O(1) payload resolution, and Nomba payload validation
│   │   └── worker/           # RabbitMQ consumers (Invoicing, State)
│   └── migrations/           # SQL migration files
│
├── frontend/                 # The Next.js Web Application
│   ├── app/                  # Next.js App Router structure
│   │   ├── (auth)/           # Merchant Login/Signup
│   │   ├── (dashboard)/      # Merchant Dashboard (Tenant)
│   │   └── pay/              # Customer Portal (Subscriber)
│   ├── components/           # Reusable UI components (Tailwind, Lucide)
│   ├── lib/                  # Utilities, TanStack React Query hooks, openapi-fetch client
│   └── mocks/                # Mock Service Worker (MSW) intercepts for offline development
│
└── nomba-docs/               # Reference documentation from the provider
```

## Authentication & Authorization

Authentication is strictly isolated between the two user personas to prevent session bleeding:

- **Merchants (Dashboard):** Use a Bearer token (`tenant_token`) passed in the `Authorization` header. This token is securely issued on login and stored in local storage, preventing third-party cookie blocking issues across domains. Authorized to manage plans, view customers, and generate portal links.
- **Subscribers (Customer Portal):** Use a Bearer token (`portal_token`) passed in the `Authorization` header. Authorized *only* to view their own invoices, update payment methods, and cancel subscriptions via magic link access.
- **API Keys (Machine-to-Machine):** Dashboard users can generate long-lived Bearer API keys. Handled securely by hashing the token in the database before storage.

## Setup & Installation

### Prerequisites
- Go 1.22+
- Node.js 18+ & npm
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3+ (or use Docker Compose)

### 1. Clone the repository
```bash
git clone https://github.com/shamigam1/subba.git
cd subba
```

### 2. Configure Environment Variables
Copy the example files and fill in your local configurations:

**Backend (`backend/.env`):**
```env
DATABASE_URL=postgres://user:pass@localhost:5432/subba
REDIS_URL=redis://localhost:6379
AMQP_URL=amqp://guest:guest@localhost:5672
NOMBA_WEBHOOK_SECRET=your_nomba_secret
APP_ENV=development
```

**Frontend (`frontend/.env.local`):**
```env
NEXT_PUBLIC_API_URL=http://localhost:8080/v1
NEXT_PUBLIC_API_MODE=live # Set to "mock" to use MSW for local UI dev without the Go backend
```

## Running Locally

### Backend
1. Apply migrations: `make migrate-up` (requires `golang-migrate`)
2. Start the API server: `go run cmd/api/main.go`
3. Start the RabbitMQ Workers: `go run cmd/worker/main.go`
4. Start the Scheduler: `go run cmd/scheduler/main.go`

### Frontend
1. Install dependencies: `npm install`
2. Start the development server: `npm run dev`
3. Access the dashboard at `http://localhost:3000`

## Database & Migrations

We use raw SQL for migrations to ensure complete control over schemas, indexes, and constraints.
- Migrations are stored in `backend/migrations`.
- We use the `github.com/golang-migrate/migrate` CLI.
- **Command:** `migrate -path ./migrations -database "$DATABASE_URL" up`

## API Overview

The Go backend exposes a robust JSON REST API, versioned at `/v1`.
- **Public:** `/v1/auth/signup`, `/v1/auth/login`, `/v1/portal/session`
- **Dashboard (Requires `subba_session`):** `/v1/plans`, `/v1/customers`, `/v1/api-keys`, `/v1/settings`
- **Portal (Requires `subba_portal`):** `/v1/portal/me`, `/v1/portal/subscription`, `/v1/portal/payment-method`
- **Webhooks:** `/v1/webhooks/nomba` (Secured via HMAC SHA-512 signature validation)

## Testing & Observability

### Testing
- **Backend:** We utilize Go's built-in `testing` package. Focus is on integration testing the RabbitMQ consumers and Postgres idempotency constraints.
- **Frontend MSW:** The frontend utilizes Mock Service Worker (`msw`) when `NEXT_PUBLIC_API_MODE="mock"`. This intercepts all network requests to allow UI development entirely offline.

### Observability
- **Prometheus & Grafana:** The Go API exposes a `/metrics` endpoint. We track:
  - `http_request_duration_seconds`
  - Database connection pool saturation
  - RabbitMQ queue depth and DLX redelivery counts

## Design Decisions & Trade-offs

- **Why TanStack Query over Redux?** We chose `@tanstack/react-query` to treat the backend as the single source of truth. Server state (plans, customers, invoices) changes rapidly via webhooks; React Query's automatic revalidation handles this seamlessly without boilerplate.
- **Why RabbitMQ over Kafka?** For a subscription engine, we do not need infinite log retention. We need explicit acknowledgment, dead-letter routing, and TTL-based retry queues for failed webhook payouts. RabbitMQ's AMQP model is perfectly suited for this.
- **Trade-off:** We skipped installing massive charting libraries (e.g., Recharts) for the dashboard overview in favor of prioritizing absolute 100% strict endpoint parity between the backend router and the frontend UI.

## Future Improvements

- **UI Polish:** Replace the `[ Recharts LineChart Placeholder ]` on the Overview page with live graphical data representations.
- **Webhooks to Tenants:** Allow merchants to register their own webhook URLs to receive events when Subba successfully charges their customers.

## License

This project is licensed under the MIT License.
