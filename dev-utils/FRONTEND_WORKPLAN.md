# Subba — Frontend Work Plan

Covers **both** frontends: the **Tenant Dashboard** (developer-facing admin) and the
**Customer Portal** (end-user, the hosted PRD slice). One Next.js app, one design
system, two audiences. Pairs with [WORKPLAN.md](WORKPLAN.md) (Track B) and the
[DESIGN_BRIEF.md](DESIGN_BRIEF.md).

## Stack

| Concern | Choice | Why |
|---|---|---|
| Framework | **Next.js (App Router) + TypeScript** | One codebase hosts both apps; server components for data fetching; trivial HTTPS hosting on Vercel → satisfies the "Working MVP URL" deliverable. |
| Styling | **Tailwind CSS** + CSS variables for tokens | Fast, consistent; tokens map 1:1 to the design brief. |
| Components | **shadcn/ui** (Radix primitives) | Accessible, unstyled-then-themed; own the code, no lock-in. |
| Data fetching | **TanStack Query** + **openapi-fetch** | Typed client generated from the frozen OpenAPI; caching, retries, loading/error states for free. |
| Types | **openapi-typescript** | Types generated from `docs/openapi.yaml` — the contract is the source of truth. |
| Mocks | **MSW** (Mock Service Worker) | Build the whole UI before the backend exists; flip to live with one env var. |
| Forms | **react-hook-form + zod** | Typed validation shared with API shapes. |
| Charts | **Recharts** | Dashboard analytics (MRR, active subs). |
| Money/dates | **Intl.NumberFormat** (`en-NG`, NGN) + **date-fns** | Correct ₦ formatting and renewal dates. |

## App architecture

A **single Next.js app** with two route groups — minimal infra for a one-week build,
shared design system, one deploy:

```
frontend/
  app/
    (dashboard)/              # tenant dashboard — served at app.<domain> (or /app)
      layout.tsx             # sidebar shell, auth guard
      login/  signup/
      overview/page.tsx       # analytics: MRR, active subs, recent payments
      plans/                  # list + create/edit (soft delete)
      customers/[id]/         # list + detail (subscription, invoices)
      api-keys/page.tsx       # generate (show once), revoke
      settings/page.tsx       # Nomba creds, sub-account, webhook URL
    (portal)/                 # customer portal — served at pay.<domain> (or /pay)
      layout.tsx              # minimal, mobile-first shell
      access/page.tsx         # passwordless entry (magic-link/token)
      page.tsx                # subscription status + next renewal (home)
      invoices/page.tsx       # invoice-history table
      payment-method/page.tsx # card on file + re-tokenization + cardless transfer
    api/                      # route handlers = BFF (httpOnly session cookies, proxy)
  components/
    ui/                       # shadcn primitives, themed
    dashboard/  portal/       # app-specific composites
  lib/
    api/                      # generated types + openapi-fetch client
    auth/                     # dashboard JWT session, portal token session
    format/                   # naira(), formatDate(), statusBadge()
    mocks/                    # MSW handlers built from OpenAPI examples
  styles/globals.css          # Tailwind layers + design tokens (CSS vars)
  tailwind.config.ts
```

Route groups are mapped to audiences by **middleware** (subdomain → group) in
production, or simple `/app` and `/pay` path prefixes for the demo. If hard
separation is ever needed, this lifts cleanly into a Turborepo with a shared `ui`
package — but don't pay that cost now.

## Two auth models (deliberately different)

- **Dashboard** — email/password → JWT issued by the Go API, stored in an **httpOnly
  cookie** set by a Next route handler (BFF pattern; the token never touches client JS).
- **Portal** — **passwordless**: the customer opens a tenant-provided link carrying a
  signed access token → exchanged for a short-lived session cookie. No password for
  end users; lowest-friction for the consumer audience.

## Mock-first strategy (the unblock)

Track B is never hard-blocked on Track A. From Day 1:
1. Generate types from the frozen `docs/openapi.yaml`.
2. MSW handlers return realistic fixtures (incl. empty/error variants).
3. `NEXT_PUBLIC_API_MODE=mock|live` toggles MSW vs. the real Go API base URL.

Build every screen against mocks, then flip to live when endpoints land (~mid-week)
by changing one env var.

## Build order & milestones (aligned to the team sprint)

| Day | Frontend deliverable |
|---|---|
| **1** | Scaffold Next app; Tailwind + tokens from design brief; shadcn primitives themed; OpenAPI → types; MSW wired; app shells for both apps render. |
| **2 (M1)** | Dashboard: auth (signup/login) + overview shell on mocks. Portal: access flow + subscription home on mocks. Design system components in place. |
| **4 (M2)** | Dashboard plans + customers CRUD on **real API**. Portal invoice-history table + **cancel flow** on real API. |
| **5 (M3)** | Portal **card re-tokenization** (Nomba widget) + **cardless transfer** panel (the demo hero). Dashboard API-keys + settings. Empty/loading/error states polished. |
| **6 (M4)** | Deploy to Vercel (HTTPS MVP URL); wire to deployed API; analytics charts; responsive + a11y pass. |
| **7 (M5)** | Demo-path polish for the cardless-renewal recording; final visual QA. |

## Screen inventory

**Tenant Dashboard** (desktop-first, data-dense)
- Signup / Login
- Overview (MRR, active subscriptions, recent payments, system health glance)
- Plans (list, create/edit, soft-delete)
- Customers (list, detail → subscription + invoices)
- API Keys (generate → show once, revoke)
- Settings (Nomba credentials, sub-account for revenue split, webhook URL + signing secret)

**Customer Portal** (mobile-first, trust-forward)
- Access (passwordless entry)
- Subscription home (plan, status badge, **next renewal date + amount**)
- Invoices (history table: date, amount, status, view)
- Payment method (card on file, **update expired card** via Nomba tokenization, **pay by
  bank transfer to virtual account** — the cardless moat, front-and-center for the demo)

## API integration notes

- Base URL via `NEXT_PUBLIC_API_BASE_URL`; all calls go through the typed
  `lib/api/client`. No `fetch` scattered in components.
- Money is `bigint` **minor units (kobo)** on the wire — format with `naira()` at the
  edge only; never do math in major units.
- Surface the backend's correlation id (`request_id`) in error toasts so a failed action
  is traceable to logs — small touch, strong judge signal.

## Quality bar

- TypeScript strict; ESLint + Prettier; `next lint` clean.
- A11y: keyboard-navigable, visible focus, WCAG AA contrast (tokens already meet it),
  semantic landmarks, labelled form fields.
- Loading (skeletons), empty, and error states are **first-class** for every data view —
  not afterthoughts.
- Lighthouse pass on the portal (it's mobile, public, and on camera).

## Deploy

- **Vercel** for both apps (one project, route groups) → instant HTTPS, satisfies the
  MVP-URL deliverable. Set `NEXT_PUBLIC_API_BASE_URL` + `NEXT_PUBLIC_API_MODE=live`.
- Backend (api/worker/scheduler + Caddy) deploys separately per the main work plan;
  frontend only needs the API's public HTTPS URL.

## Dependencies on Track A

- **`docs/openapi.yaml` frozen Day 1** — the single thing that unblocks all of this.
- Nomba **tokenization** flow details (hosted widget vs. inline SDK) for the card-update
  screen — confirm with Track A which Nomba integration the portal calls.
- Customer **access-token** format for the passwordless portal link.
