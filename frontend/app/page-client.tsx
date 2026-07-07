"use client";

import Link from "next/link";
import { motion } from "motion/react";
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
import { SubbaLogo } from "@/components/brand/subba-logo";

export function LandingClient() {
  return (
    <div className="flex min-h-dvh flex-col bg-slate-50 text-slate-900 relative overflow-hidden">
      {/* Floating Background Icons & Doodles */}
      <motion.div
        className="absolute top-20 left-10 text-slate-300/20"
        animate={{ y: [0, 20, 0], rotate: [0, 10, -10, 0] }}
        transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }}
      >
        <Activity size={80} />
      </motion.div>
      <motion.div
        className="absolute bottom-40 right-20 text-slate-300/20"
        animate={{ y: [0, -30, 0], rotate: [0, -15, 15, 0] }}
        transition={{ duration: 8, repeat: Infinity, ease: "easeInOut", delay: 1 }}
      >
        <Lock size={100} />
      </motion.div>

      {/* Doodles */}
      <motion.svg
        className="absolute top-1/4 right-32 text-slate-200/30 w-32 h-32"
        viewBox="0 0 100 100"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        animate={{ x: [0, -20, 0], y: [0, 10, 0] }}
        transition={{ duration: 10, repeat: Infinity, ease: "easeInOut" }}
      >
        <path d="M10 50 Q 25 25 50 50 T 90 50" />
        <path d="M10 60 Q 25 35 50 60 T 90 60" />
      </motion.svg>
      <motion.div
        className="absolute bottom-20 left-1/4 w-16 h-16 border-4 border-slate-200/30 rounded-full"
        animate={{ scale: [1, 1.2, 1], opacity: [0.3, 0.1, 0.3] }}
        transition={{ duration: 5, repeat: Infinity, ease: "easeInOut", delay: 2 }}
      />

      {/* Nav */}
      <header className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-5 relative z-10">
        <SubbaLogo size="sm" showText={true} />
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
      <section className="mx-auto grid w-full max-w-6xl items-center gap-12 px-6 py-16 sm:py-24 lg:grid-cols-2 relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
        >
          <p className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white px-3 py-1 text-xs font-medium text-slate-600 shadow-sm">
            <span className="h-1.5 w-1.5 rounded-full bg-brand-600" aria-hidden />
            Built on Nomba payment rails
          </p>
          <h1 className="mt-5 text-5xl font-extrabold leading-tight tracking-tight sm:text-6xl text-slate-900">
            Recurring billing that never drops a payment
          </h1>
          <p className="mt-5 max-w-lg text-lg leading-8 text-slate-600">
            Subba is a plug-and-play subscriptions engine for African SaaS —
            cardless renewals by bank transfer, automatic revenue splits, and a
            fault-tolerant billing pipeline. Naira-native, mobile-first.
          </p>
          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <Button asChild size="lg" className="bg-brand-600 hover:bg-brand-700">
              <Link href="/signup">
                Start billing
                <ArrowRight className="h-4 w-4 ml-2" aria-hidden />
              </Link>
            </Button>
            <Button asChild size="lg" variant="outline" className="bg-white">
              <Link href="/pay/access">See the customer portal</Link>
            </Button>
          </div>
        </motion.div>

        {/* Mock portal card — sells the end-user experience at a glance. */}
        <motion.div
          className="mx-auto w-full max-w-sm"
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-xl shadow-slate-200/50">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-slate-500 font-medium">Premium Membership</p>
                <p className="mt-1 text-4xl font-bold tabular-nums tracking-tight text-slate-900">
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
            <div className="mt-5 rounded-xl border border-dashed border-slate-300 bg-slate-50 p-5">
              <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                Renew by bank transfer
              </p>
              <p className="mt-1 text-2xl font-bold tabular-nums tracking-wide text-slate-900">
                012 345 6789
              </p>
              <p className="mt-3 inline-flex items-center gap-1.5 text-xs font-medium text-green-700">
                <RefreshCw className="h-3.5 w-3.5" aria-hidden />
                Transfer detected — subscription renewed
              </p>
            </div>
          </div>
          <p className="mt-4 text-center text-sm text-slate-500 font-medium">
            What your customers see — no card required.
          </p>
        </motion.div>
      </section>

      {/* Developer API Section */}
      <section className="mx-auto w-full max-w-6xl px-6 py-24 relative z-10">
        <motion.div
          className="rounded-3xl bg-slate-900 p-8 sm:p-12 shadow-2xl flex flex-col lg:flex-row gap-12 items-center"
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.7 }}
        >
          <div className="lg:w-1/2">
            <h2 className="text-3xl font-bold text-white mb-4">
              Developer-first API
            </h2>
            <p className="text-slate-400 text-lg mb-8 leading-relaxed">
              Integrate Subba into your product in minutes. Create plans, manage customers, and provision subscriptions with a single API call. We handle the webhooks, retries, and edge cases.
            </p>
            <ul className="space-y-4">
              {[
                "Idempotent requests",
                "Standard HMAC-SHA256 Webhooks",
                "Predictable error formats",
              ].map((item, i) => (
                <li key={i} className="flex items-center gap-3 text-slate-300">
                  <CheckCircle2 className="h-5 w-5 text-brand-400" />
                  {item}
                </li>
              ))}
            </ul>
          </div>
          <div className="lg:w-1/2 w-full">
            <div className="rounded-xl bg-slate-950 border border-slate-800 p-4 font-mono text-sm overflow-x-auto shadow-inner">
              <div className="flex gap-2 mb-4">
                <div className="w-3 h-3 rounded-full bg-red-500/20 border border-red-500/50" />
                <div className="w-3 h-3 rounded-full bg-yellow-500/20 border border-yellow-500/50" />
                <div className="w-3 h-3 rounded-full bg-green-500/20 border border-green-500/50" />
              </div>
              <pre className="text-slate-300">
                <code className="language-js">
{`const res = await fetch("https://api.subba.com/v1/subscriptions", {
  method: "POST",
  headers: {
    "Authorization": "Bearer sk_test_...",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    customer_id: "cus_9f8b7c...",
    plan_id: "plan_monthly_premium",
    start_date: "2026-08-01T00:00:00Z"
  })
});

const sub = await res.json();
console.log(sub.status); // "active"`}
                </code>
              </pre>
            </div>
          </div>
        </motion.div>
      </section>

      {/* Feature Grid */}
      <section id="features" className="mx-auto w-full max-w-6xl px-6 py-24 relative z-10">
        <div className="text-center max-w-2xl mx-auto mb-16">
          <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">Everything you need to scale</h2>
          <p className="mt-4 text-lg text-slate-600">We abstracted away the hardest parts of recurring billing in emerging markets so you can focus on building your product.</p>
        </div>
        
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-8">
          {[
            {
              icon: Landmark,
              title: "Naira-Native Engine",
              desc: "Built specifically for the Nigerian market with deep Nomba integration.",
            },
            {
              icon: Split,
              title: "Automated Splits",
              desc: "Split subscription revenue instantly across multiple sub-accounts.",
            },
            {
              icon: Smartphone,
              title: "Bank Transfer Renewals",
              desc: "A frictionless portal for customers who prefer transferring over cards.",
            },
            {
              icon: Activity,
              title: "Fault-Tolerant Webhooks",
              desc: "Built on RabbitMQ. If a webhook drops, we queue and retry automatically.",
            },
            {
              icon: Lock,
              title: "Enterprise Security",
              desc: "HMAC signatures, isolated tenant execution, and encrypted secrets.",
            },
            {
              icon: RefreshCw,
              title: "Smart Dunning",
              desc: "Intelligent retry schedules and grace periods to recover failed payments.",
            },
          ].map((feat, i) => (
            <motion.div
              key={i}
              className="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm hover:shadow-md transition-shadow"
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
            >
              <div className="w-12 h-12 rounded-lg bg-brand-50 flex items-center justify-center text-brand-600 mb-4">
                <feat.icon size={24} />
              </div>
              <h3 className="text-xl font-bold mb-2">{feat.title}</h3>
              <p className="text-slate-600 leading-relaxed">{feat.desc}</p>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-slate-200 bg-white relative z-10">
        <div className="mx-auto w-full max-w-6xl px-6 py-12 md:flex md:items-center md:justify-between">
          <div className="flex justify-center md:justify-start">
            <SubbaLogo size="sm" showText={true} />
          </div>
          <div className="mt-8 md:mt-0">
            <p className="text-center text-sm leading-5 text-slate-500">
              &copy; {new Date().getFullYear()} Subba Inc. All rights reserved.
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}
