import Link from "next/link";
import {
  Activity,
  ArrowRight,
  CheckCircle2,
  Landmark,
  Lock,
  RefreshCw,
  Smartphone,
  Split,
} from "lucide-react";

import { Button } from "@/components/ui/button";

export default function LandingPage() {
  return (
    <div className="flex min-h-dvh flex-col bg-white text-slate-900">
      {/* Nav */}
      <header className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-5">
        <p className="text-xl font-bold tracking-tight">
          subba<span className="text-primary">.</span>
        </p>
        <nav className="hidden items-center gap-6 text-sm font-medium text-slate-600 sm:flex">
          <a href="#features" className="hover:text-slate-900">
            Features
          </a>
          <a href="#how-it-works" className="hover:text-slate-900">
            How it works
          </a>
          <Link href="/pay/access" className="hover:text-slate-900">
            Customer portal
          </Link>
        </nav>
        <Button asChild>
          <Link href="/signup">Open dashboard</Link>
        </Button>
      </header>

      {/* Hero */}
      <section className="mx-auto grid w-full max-w-6xl items-center gap-12 px-6 py-16 sm:py-24 lg:grid-cols-2">
        <div>
          <p className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-600">
            <span className="h-1.5 w-1.5 rounded-full bg-primary" aria-hidden />
            Built on Nomba payment rails
          </p>
          <h1 className="mt-5 text-4xl font-bold leading-tight tracking-tight sm:text-5xl">
            Recurring billing that never drops a payment
          </h1>
          <p className="mt-5 max-w-lg text-lg leading-8 text-slate-600">
            Subba is a plug-and-play subscriptions engine for African SaaS —
            cardless renewals by bank transfer, automatic revenue splits, and a
            fault-tolerant billing pipeline. Naira-native, mobile-first.
          </p>
          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <Button asChild size="lg">
              <Link href="/signup">
                Start billing
                <ArrowRight className="h-4 w-4" aria-hidden />
              </Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href="/pay/access">See the customer portal</Link>
            </Button>
          </div>
        </div>

        {/* Mock portal card — sells the end-user experience at a glance. */}
        <div className="mx-auto w-full max-w-sm">
          <div className="rounded-2xl border bg-white p-6 shadow-sm">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-slate-500">Premium Membership</p>
                <p className="mt-1 text-3xl font-semibold tabular-nums tracking-tight">
                  ₦15,000
                  <span className="text-base font-normal text-slate-500">/mo</span>
                </p>
              </div>
              <span className="inline-flex items-center gap-1.5 rounded-full border border-green-200 bg-green-50 px-2.5 py-1 text-xs font-medium text-green-700">
                <CheckCircle2 className="h-3.5 w-3.5" aria-hidden />
                Active
              </span>
            </div>
            <p className="mt-4 text-sm text-slate-500">Renews 30 Jul 2026</p>
            <div className="mt-5 rounded-xl border border-dashed border-slate-200 bg-slate-50 p-4">
              <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                Renew by bank transfer
              </p>
              <p className="mt-1 text-lg font-semibold tabular-nums tracking-wide">
                012 345 6789
              </p>
              <p className="mt-2 inline-flex items-center gap-1.5 text-xs text-green-700">
                <RefreshCw className="h-3.5 w-3.5" aria-hidden />
                Transfer detected — subscription renewed
              </p>
            </div>
          </div>
          <p className="mt-3 text-center text-xs text-slate-400">
            What your customers see — no card required.
          </p>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="border-t bg-slate-50">
        <div className="mx-auto w-full max-w-6xl px-6 py-16 sm:py-20">
          <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Billing infrastructure, not billing chores
          </h2>
          <p className="mt-3 max-w-2xl text-slate-600">
            Everything between &ldquo;customer subscribed&rdquo; and &ldquo;money in your
            account&rdquo; — handled, retried, and observable.
          </p>
          <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {features.map(({ icon: Icon, title, body }) => (
              <div key={title} className="rounded-xl border bg-white p-6">
                <div className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                  <Icon className="h-5 w-5 text-primary" aria-hidden />
                </div>
                <h3 className="mt-4 font-semibold">{title}</h3>
                <p className="mt-2 text-sm leading-6 text-slate-600">{body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* How it works */}
      <section id="how-it-works" className="mx-auto w-full max-w-6xl px-6 py-16 sm:py-20">
        <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
          Live in an afternoon
        </h2>
        <ol className="mt-10 grid gap-8 sm:grid-cols-3">
          {steps.map(({ title, body }, i) => (
            <li key={title} className="relative">
              <p className="text-sm font-semibold text-primary">Step {i + 1}</p>
              <h3 className="mt-2 font-semibold">{title}</h3>
              <p className="mt-2 text-sm leading-6 text-slate-600">{body}</p>
            </li>
          ))}
        </ol>
        <div className="mt-12 rounded-2xl bg-primary px-8 py-10 text-center sm:py-12">
          <h2 className="text-2xl font-bold tracking-tight text-white sm:text-3xl">
            Start collecting subscriptions today
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-primary-foreground/80">
            Create your plans, share your portal link, and let renewals run
            themselves.
          </p>
          <Button asChild size="lg" variant="secondary" className="mt-6">
            <Link href="/signup">
              Open the dashboard
              <ArrowRight className="h-4 w-4" aria-hidden />
            </Link>
          </Button>
        </div>
      </section>

      {/* Footer */}
      <footer className="mt-auto border-t">
        <div className="mx-auto flex w-full max-w-6xl flex-col items-center justify-between gap-3 px-6 py-8 text-sm text-slate-500 sm:flex-row">
          <p>
            subba<span className="text-primary">.</span> — Nomba Subscriptions Engine
          </p>
          <p>Nomba × DevCareer Hackathon 2026 · Infrastructure Track</p>
        </div>
      </footer>
    </div>
  );
}

const features = [
  {
    icon: Landmark,
    title: "Cardless renewal",
    body: "Every customer gets a personal virtual account. They renew with a plain bank transfer — we detect it and advance the subscription automatically. No debit card needed.",
  },
  {
    icon: Split,
    title: "Automatic revenue split",
    body: "Your share of every payment moves to your Nomba sub-account the moment it lands. No manual settlement, no spreadsheet reconciliation.",
  },
  {
    icon: RefreshCw,
    title: "Fault-tolerant pipeline",
    body: "Payment events are queued, processed exactly once, and retried with backoff when anything downstream fails. A dead-letter queue catches the rest — nothing is silently lost.",
  },
  {
    icon: Smartphone,
    title: "Hosted customer portal",
    body: "A passwordless, mobile-first portal your customers open from a magic link — invoices, cancellation, and payment methods, co-branded as yours.",
  },
  {
    icon: Lock,
    title: "Tenant isolation by default",
    body: "Row-level security in Postgres walls every tenant off at the database layer. API keys are hashed, secrets encrypted, sessions revocable.",
  },
  {
    icon: Activity,
    title: "Observable from day one",
    body: "Prometheus metrics and a Grafana dashboard ship in the box — consumer throughput, queue depth, latency percentiles, and payment success rates.",
  },
];

const steps = [
  {
    title: "Create your plans",
    body: "Sign up, define plans in naira, and add customers — from the dashboard or the API.",
  },
  {
    title: "Share your portal link",
    body: "Customers get a secure magic link to subscribe, pay by transfer or card, and manage everything themselves.",
  },
  {
    title: "Renewals run themselves",
    body: "Subba invoices, collects, splits revenue to your sub-account, and chases failed payments with smart retries.",
  },
];
