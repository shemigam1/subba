# Subba Frontend

This is the Next.js App Router frontend for the **Subba** subscription engine.

## Live Deployment
- **Live Frontend (Vercel):** https://subba-theta.vercel.app

## Overview

The frontend is divided into two primary isolated experiences:
1. **The Merchant Dashboard (`/`)**: A secure, multi-tenant portal where SaaS developers can create recurring billing plans, manage their customers' subscriptions, and view analytics. It utilizes local storage Bearer tokens for secure, cross-site authentication with the Go backend.
2. **The Customer Portal (`/pay`)**: A passwordless, magic-link interface where end-users can securely view their invoices, update payment methods, and cancel subscriptions.

## Tech Stack
- **Framework:** Next.js 14 (App Router)
- **Data Fetching:** `@tanstack/react-query` (with `openapi-fetch` for strict types)
- **Styling:** Tailwind CSS + Shadcn UI
- **Animations:** Framer Motion

## Getting Started

First, install dependencies:
```bash
npm install
```

Then, run the development server:
```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the application.

## API Mocking (Offline Mode)
To work on the UI without running the Go backend, set `NEXT_PUBLIC_API_MODE="mock"` in your `.env.local` file. This enables Mock Service Worker (MSW) to intercept all API calls and return mock data.

