"use client";

import { useState } from "react";
import { Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { naira } from "@/lib/format";
import type { components } from "@/lib/api/v1";

type Plan = components["schemas"]["Plan"];

export default function PlansPage() {
  const queryClient = useQueryClient();
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  
  // Form state
  const [name, setName] = useState("");
  const [amount, setAmount] = useState("");
  const [interval, setInterval] = useState<"month" | "year">("month");

  // Data Fetching
  const { data: plans = [], isLoading } = useQuery({
    queryKey: ["plans"],
    queryFn: async () => {
      const { data, error } = await api.GET("/plans");
      if (error) throw error;
      return data as Plan[];
    },
  });

  // Data Mutation
  const createPlan = useMutation({
    mutationFn: async () => {
      const amountMinor = Math.round(parseFloat(amount) * 100);
      const { data, error } = await api.POST("/plans", {
        body: {
          name,
          amount_minor: amountMinor,
          currency: "NGN",
          interval,
        },
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["plans"] });
      setIsDrawerOpen(false);
      setName("");
      setAmount("");
      setInterval("month");
    },
  });

  function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    createPlan.mutate();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">Plans</h1>
          <p className="text-sm text-slate-500">
            Manage your subscription tiers and pricing.
          </p>
        </div>
        <Button onClick={() => setIsDrawerOpen(true)} className="gap-2">
          <Plus className="h-4 w-4" /> New Plan
        </Button>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center text-slate-500 animate-pulse">Loading plans...</div>
        ) : plans.length === 0 ? (
          <div className="p-12 text-center text-slate-500 flex flex-col items-center">
            <div className="w-12 h-12 bg-slate-100 rounded-full flex items-center justify-center mb-4">
              <Plus className="w-6 h-6 text-slate-400" />
            </div>
            <h3 className="font-semibold text-slate-900">No plans yet</h3>
            <p className="text-sm mt-1">Create your first plan to start charging customers.</p>
            <Button variant="outline" className="mt-4" onClick={() => setIsDrawerOpen(true)}>Create Plan</Button>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-600">
              <thead className="bg-slate-50 border-b border-slate-200 text-slate-500 font-medium">
                <tr>
                  <th className="px-6 py-4">Name</th>
                  <th className="px-6 py-4">Amount</th>
                  <th className="px-6 py-4">Interval</th>
                  <th className="px-6 py-4">Status</th>
                  <th className="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {plans.map((plan) => (
                  <tr key={plan.id} className="hover:bg-slate-50 transition-colors">
                    <td className="px-6 py-4 font-medium text-slate-900">{plan.name}</td>
                    <td className="px-6 py-4 tabular-nums">{naira(plan.amount_minor || 0)}</td>
                    <td className="px-6 py-4 capitalize">{plan.interval}</td>
                    <td className="px-6 py-4">
                      {plan.deleted_at ? (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-slate-100 text-slate-600">
                          Archived
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-success-600/10 text-success-600">
                          Active
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <Button variant="ghost" size="sm" className="text-brand-600 hover:text-brand-700">
                        Edit
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Drawer Overlay */}
      {isDrawerOpen && (
        <div className="fixed inset-0 z-50 flex justify-end bg-slate-900/50 backdrop-blur-sm transition-opacity">
          <div className="w-full max-w-md bg-white h-full shadow-2xl flex flex-col animate-in slide-in-from-right duration-300">
            <div className="flex items-center justify-between p-6 border-b border-slate-100">
              <h2 className="text-lg font-bold text-slate-900">Create New Plan</h2>
              <button onClick={() => setIsDrawerOpen(false)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-6 flex-1 overflow-y-auto">
              <form id="create-plan-form" onSubmit={handleCreate} className="space-y-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium text-slate-900">Plan Name</label>
                  <Input 
                    required 
                    placeholder="e.g. Pro Tier" 
                    value={name} 
                    onChange={(e) => setName(e.target.value)} 
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-slate-900">Amount (₦)</label>
                  <Input 
                    required 
                    type="number"
                    min="1"
                    placeholder="5000" 
                    value={amount} 
                    onChange={(e) => setAmount(e.target.value)} 
                  />
                  <p className="text-xs text-slate-500">Enter the amount in major units (Naira).</p>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-slate-900">Billing Interval</label>
                  <select 
                    className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm ring-offset-white file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-slate-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    value={interval}
                    onChange={(e) => setInterval(e.target.value as "month" | "year")}
                  >
                    <option value="month">Monthly</option>
                    <option value="year">Yearly</option>
                  </select>
                </div>
              </form>
            </div>
            <div className="p-6 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
              <Button type="button" variant="ghost" onClick={() => setIsDrawerOpen(false)}>Cancel</Button>
              <Button type="submit" form="create-plan-form" disabled={createPlan.isPending}>
                {createPlan.isPending ? "Creating..." : "Create Plan"}
              </Button>
            </div>
            {createPlan.isError && (
              <div className="p-4 bg-red-50 text-red-700 text-sm">Failed to create plan.</div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

