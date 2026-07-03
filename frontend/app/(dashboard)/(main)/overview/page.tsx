"use client";

import { Activity, CreditCard, DollarSign, Users, AlertCircle } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { naira } from "@/lib/format";
import type { components } from "@/lib/api/v1";

type AnalyticsOverview = components["schemas"]["AnalyticsOverview"];

export default function OverviewPage() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["analytics", "overview"],
    queryFn: async () => {
      const { data, error } = await api.GET("/analytics/overview");
      if (error) throw error;
      return data as AnalyticsOverview;
    },
  });

  if (isLoading) {
    return (
      <div className="space-y-8 animate-pulse">
        <div className="h-8 w-48 bg-slate-200 rounded"></div>
        <div className="h-20 w-full bg-slate-200 rounded-xl"></div>
        <div className="grid gap-4 md:grid-cols-3">
          <div className="h-32 bg-slate-200 rounded-xl"></div>
          <div className="h-32 bg-slate-200 rounded-xl"></div>
          <div className="h-32 bg-slate-200 rounded-xl"></div>
        </div>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="p-6 rounded-xl border border-red-200 bg-red-50 text-red-700">
        <h3 className="font-semibold text-red-900">Failed to load analytics</h3>
        <p className="text-sm mt-1">Please try again later or check your connection.</p>
      </div>
    );
  }

  const isHealthy = data.dlq_depth === 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-900">Overview</h1>
        <p className="text-sm text-slate-500">
          Your revenue and system health at a glance.
        </p>
      </div>

      {/* System Health Strip */}
      <div className={`rounded-xl border p-4 flex items-center justify-between ${isHealthy ? 'bg-success-600/10 border-success-600/20' : 'bg-danger-600/10 border-danger-600/20'}`}>
        <div className="flex items-center gap-3">
          {isHealthy ? (
            <Activity className="h-5 w-5 text-success-600" />
          ) : (
            <AlertCircle className="h-5 w-5 text-danger-600" />
          )}
          <div>
            <h3 className={`text-sm font-semibold ${isHealthy ? 'text-success-600' : 'text-danger-600'}`}>
              System Health: {isHealthy ? 'Healthy' : 'Degraded'}
            </h3>
            <p className={`text-xs mt-0.5 ${isHealthy ? 'text-success-600/80' : 'text-danger-600/80'}`}>
              {data.dlq_depth} failed payments in Dead Letter Queue
            </p>
          </div>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {/* Stat Card 1 */}
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-row items-center justify-between pb-2">
            <h3 className="tracking-tight text-sm font-medium text-slate-500">
              Monthly Recurring Revenue
            </h3>
            <DollarSign className="h-4 w-4 text-slate-400" />
          </div>
          <div className="text-2xl font-bold text-slate-900 tabular-nums">
            {data.mrr ? naira(data.mrr.amount_minor) : "₦0.00"}
          </div>
          <p className="text-xs text-slate-500 mt-1">Based on active plans</p>
        </div>

        {/* Stat Card 2 */}
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-row items-center justify-between pb-2">
            <h3 className="tracking-tight text-sm font-medium text-slate-500">
              Active Subscriptions
            </h3>
            <Users className="h-4 w-4 text-slate-400" />
          </div>
          <div className="text-2xl font-bold text-slate-900 tabular-nums">
            {data.active_subscriptions}
          </div>
          <p className="text-xs text-slate-500 mt-1">Currently subscribed customers</p>
        </div>

        {/* Stat Card 3 */}
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-row items-center justify-between pb-2">
            <h3 className="tracking-tight text-sm font-medium text-slate-500">
              Payments Today
            </h3>
            <CreditCard className="h-4 w-4 text-slate-400" />
          </div>
          <div className="text-2xl font-bold text-slate-900 tabular-nums">
            {data.payments_today}
          </div>
          <p className="text-xs text-slate-500 mt-1">{data.failed_payments} failed payments</p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        {/* Revenue Chart Placeholder */}
        <div className="col-span-4 rounded-xl border border-slate-200 bg-white shadow-sm flex flex-col overflow-hidden">
          <div className="p-6 border-b border-slate-100">
            <h3 className="font-semibold leading-none tracking-tight text-slate-900">Revenue Over Time</h3>
            <p className="text-sm text-slate-500 mt-2">Daily revenue for the last 30 days.</p>
          </div>
          <div className="p-6 flex-1 flex items-center justify-center min-h-[300px] text-slate-400 bg-slate-50/50">
            [ Recharts LineChart Placeholder - Data ready in state ]
          </div>
        </div>

        {/* Recent Payments Placeholder */}
        <div className="col-span-3 rounded-xl border border-slate-200 bg-white shadow-sm flex flex-col overflow-hidden">
          <div className="p-6 border-b border-slate-100">
            <h3 className="font-semibold leading-none tracking-tight text-slate-900">Recent Payments</h3>
            <p className="text-sm text-slate-500 mt-2">Latest transactions across all customers.</p>
          </div>
          <div className="p-6 flex-1 flex items-center justify-center min-h-[300px] text-slate-400 bg-slate-50/50">
            [ Recent Payments Table Placeholder ]
          </div>
        </div>
      </div>
    </div>
  );
}

