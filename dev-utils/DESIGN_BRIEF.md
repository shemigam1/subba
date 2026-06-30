# Subba — Design Brief

A brief for generating UI designs/mockups. Subba is a **managed subscriptions &
recurring-billing engine** on Nomba's African-first payment rails. There are **two
products** sharing **one design system**:

1. **Tenant Dashboard** — for SaaS developers/businesses who run billing on Subba.
   Data-dense, professional, desktop-first. Feels like a precise developer tool.
2. **Customer Portal** — hosted, public, for *their* end-users to manage a
   subscription. Trust-forward, friendly, **mobile-first** (Nigerian end-users are
   mobile-heavy).

> If you only design one hero flow, design the **cardless renewal** in the portal:
> an end-user with no debit card pays by **bank transfer to a virtual account** and
> their subscription auto-renews. It is the product's moat and the demo centrepiece.

---

## Audiences & emotional goals

| | Tenant Dashboard | Customer Portal |
|---|---|---|
| Who | Developer / finance ops at a SaaS | A regular person paying for a service |
| Device | Desktop primarily | **Phone primarily** |
| Feeling to evoke | "This is precise and in control. I trust it with money." | "This is simple and safe. I know exactly what I'm paying and when." |
| Anti-goal | Toy-like, vague, cluttered | Corporate, intimidating, jargon |

## Design principles

1. **Money is sacred.** Amounts, dates, and statuses are the most legible things on
   every screen. Tabular numerals, clear ₦ formatting, no ambiguity about *what* and *when*.
2. **State is always visible.** Subscription status, invoice status, system health —
   shown as unmistakable badges with color + label + icon (never color alone).
3. **Calm density, not noise.** Dashboard is dense but breathable; portal is generous
   and focused — one primary action per screen.
4. **Trust through clarity.** Plain language ("Your next payment", not "Upcoming
   dunning event"). Security cues where money moves.
5. **African-first.** Naira-native, bank-transfer as a first-class payment path (not an
   afterthought to cards), mobile-first.

---

## Visual direction

**Mood:** modern fintech infrastructure — confident, clean, a little bold. Reference
points: the clarity of Stripe's dashboard + the warmth/approachability a consumer
Nigerian fintech (Paystack/Kuda-era) brings to end-user screens. Not flashy; trustworthy.

### Color tokens

Brand violet conveys trust + a distinct, non-generic fintech identity. Money-green for
positive/active states. (Swap brand hue if Nomba co-brand colors are required.)

| Token | Hex | Use |
|---|---|---|
| `--brand-700` | `#4A39C4` | hover/pressed primary |
| `--brand-600` | `#5B47E0` | **primary** buttons, active nav, links |
| `--brand-500` | `#6D5CE7` | accents |
| `--brand-50`  | `#EEEBFC` | tinted surfaces, selected rows |
| `--success-600` | `#16A34A` | paid, active, funded |
| `--warning-600` | `#D97706` | past_due, expiring card, action needed |
| `--danger-600`  | `#DC2626` | failed, canceled, DLQ |
| `--info-600`    | `#2563EB` | informational, in-progress |
| `--slate-900` | `#0F172A` | primary text, dashboard dark surface |
| `--slate-700` | `#334155` | secondary text |
| `--slate-500` | `#64748B` | muted/labels |
| `--slate-200` | `#E2E8F0` | borders |
| `--slate-50`  | `#F8FAFC` | app background |
| `--white`     | `#FFFFFF` | cards |

All text/background pairs meet **WCAG AA**. Status is **never color-only** — always
color + label + icon.

### Typography
- **Inter** for everything (UI + body). Optional display face (e.g. *Clash Display* or
  *Geist*) for portal marketing headlines only.
- Money & tables use **tabular-nums**.
- Scale: 12 / 14 (base) / 16 / 20 / 24 / 32 / 40. Dashboard base 14, portal base 16.

### Shape, space, elevation
- **Spacing:** 4px base scale.
- **Radius:** inputs 8px, cards 12px, modals 16px, pills/badges full.
- **Elevation:** mostly flat with hairline `--slate-200` borders; soft shadow only on
  overlays/menus. Avoid heavy drop shadows.
- **Icons:** Lucide, 1.5px stroke, 20px default.
- **Density:** dashboard table rows ~44px; portal comfortable, large tap targets (≥44px).

### Currency & locale
- Always ₦ with thousands separators: **₦5,000** (from `5000` major; wire format is
  kobo/minor units — design shows major). Intervals: "/mo", "/yr".
- Dates as "30 Jun 2026"; renewal phrased "Renews 30 Jun 2026".

---

## Shared component inventory

Buttons (primary/secondary/ghost/danger) · Inputs, Select, Textarea, Switch · **Status
Badge** (status → color+icon+label map below) · **Money cell** (right-aligned, tabular) ·
Data Table (sortable, paginated, sticky header) · Card / Stat card · Tabs · Modal /
Drawer · Toast (success/error, includes copyable `request_id`) · Empty state · Skeleton
loader · Inline error · Avatar / tenant logo · Copy-to-clipboard field (API keys, account no.).

**Status → visual map**
- Subscription: `active`→success · `past_due`→warning · `incomplete`→info ·
  `canceled`→slate · `unpaid`→danger
- Invoice: `paid`→success · `open`→info · `void`→slate · `uncollectible`→danger
- System/health (dashboard glance): healthy→success · DLQ rising→danger

---

## Tenant Dashboard — screens

Shell: left sidebar (logo, nav: Overview · Plans · Customers · API Keys · Settings),
top bar (tenant name, environment switch live/test, user menu). Desktop-first; sidebar
collapses on tablet.

1. **Signup / Login** — centered card, email + password, brand panel on the side with a
   one-line value prop. Clear error states.
2. **Overview** — 4 stat cards (MRR, Active subscriptions, Payments today, Failed/DLQ),
   a revenue line chart (last 30 days), a recent-payments table, and a small **system
   health** strip (consumer health, DLQ depth) — a subtle nod to the fault-tolerance story.
3. **Plans** — table (name, amount, interval, status). "New plan" opens a drawer:
   name, amount (₦), interval (monthly/yearly), currency. Soft-deleted plans shown muted
   with a "Restore" affordance.
4. **Customers** — searchable table (name, email, active plan, status). Row → **Customer
   detail**: profile, current subscription (status badge, period dates), invoice history
   sub-table, payment method on file.
5. **API Keys** — generate key → **one-time reveal modal** with copy + "store it now"
   warning; list shows masked keys with created date + revoke. Emphasize "shown once".
6. **Settings** — Nomba credentials (client id/secret, masked), **sub-account for revenue
   split**, **webhook URL + signing secret** (copy field), with a "we never log your
   secret" reassurance.

## Customer Portal — screens

Shell: minimal top bar (tenant's logo — co-branded, not Subba-branded), generous mobile
layout, one clear primary action per screen.

1. **Access** — passwordless: "Enter the email your subscription is under" → "Check your
   inbox" confirmation. Calm, reassuring, single field. (Or direct token link → straight in.)
2. **Subscription home (hero)** — big, legible card: plan name, **status badge**, the
   amount, and **"Renews 30 Jun 2026"**. Primary actions: *Update payment method*,
   *Cancel subscription*. If `past_due`: a warning banner with **"Pay now by bank
   transfer"** leading to the cardless flow.
3. **Invoices** — clean history table/list (date, amount, status, View). Mobile = stacked
   cards; desktop = table. Each row → invoice detail with line items (incl. proration).
4. **Payment method** —
   - **Card on file**: brand + last4 + expiry; **"Update card"** opens the **Nomba
     tokenization** widget; if expired, a warning state nudges the update.
   - **Pay by bank transfer (cardless — the moat & demo hero):** show the customer's
     **virtual account number** (bank, account no., copy button) with "Transfer ₦5,000 to
     this account to renew. We detect it automatically." A subtle live "waiting for
     transfer → detected → renewed" state sells the magic on camera.
5. **Cancel flow** — confirmation modal: what they keep until period end ("Active until
   30 Jun 2026"), a soft retention line, confirm/keep. Post-cancel state shows `canceled`
   badge + "Resubscribe".

---

## States & edge cases (design all of these)

For every data screen provide: **loading** (skeleton, not spinner-only), **empty**
(friendly, with the primary CTA), **error** (inline, retry, shows `request_id` for
support), and **success** (toast/inline confirmation). Specific ones to nail:
- Portal `past_due` → prominent but non-alarming "renew by transfer" path.
- Card **expiring/expired** → warning on payment method + subscription home.
- API key reveal → unmistakable "you will only see this once".
- Cardless **waiting → detected → renewed** micro-states on the transfer panel.

## Responsive
- **Portal: mobile-first** (360–414px primary), scaling up to a centered ~560px column
  on desktop. Tap targets ≥44px.
- **Dashboard: desktop-first** (1280px), graceful down to tablet (sidebar → drawer);
  tables become horizontally scrollable or stacked on small screens.

## Accessibility
WCAG AA contrast (tokens comply), full keyboard nav, visible focus rings (brand-600,
2px), status conveyed by icon+label not color alone, labelled inputs, `prefers-reduced-motion`
respected on the cardless live-state animation.

---

## Deliverables requested from design

1. **Design tokens** (color, type, spacing, radius) as a usable spec/Tailwind theme.
2. **Core component sheet** (buttons, inputs, badges, table, stat card, modal, empty/loading).
3. **Tenant Dashboard**: hi-fi mockups for Overview, Plans (+ create drawer), Customer
   detail, API Keys (reveal modal), Settings. Desktop.
4. **Customer Portal**: hi-fi mockups for Subscription home, Invoices, Payment method
   (incl. **cardless transfer panel** + its waiting/detected/renewed states), Cancel modal.
   Mobile + desktop.
5. **The hero**: a polished cardless-renewal sequence (portal) suitable for the demo video.
