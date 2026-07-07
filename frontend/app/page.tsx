import type { Metadata } from "next";
import { LandingClient } from "./page-client";

export const metadata: Metadata = {
  title: "Subba | Smarter subscriptions for African businesses",
  description: "Subba is a plug-and-play subscriptions engine for African SaaS — cardless renewals by bank transfer, automatic revenue splits, and a fault-tolerant billing pipeline.",
  alternates: {
    canonical: "/",
  },
};

export default function LandingPage() {
  return <LandingClient />;
}
