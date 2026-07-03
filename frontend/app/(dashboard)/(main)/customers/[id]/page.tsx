"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowLeft, User, CreditCard, Link as LinkIcon, Loader2, Plus, X, Trash2, Pencil } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { naira, formatDate } from "@/lib/format";
import type { components } from "@/lib/api/v1";

type Customer = components["schemas"]["Customer"];
type Subscription = components["schemas"]["Subscription"];
type Plan = components["schemas"]["Plan"];

interface Invoice {
  id: string;
  amount_minor: number;
  status: string;
  created_at: string;
  invoice_url?: string;
}

export default function CustomerDetailPage({ params }: { params: { id: string } }) {
  const customerId = params.id;
  const queryClient = useQueryClient();
  const [portalLink, setPortalLink] = useState("");
  
  // Drawer states
  const [isSubDrawerOpen, setIsSubDrawerOpen] = useState(false);
  const [selectedPlanId, setSelectedPlanId] = useState("");
  const [isEditDrawerOpen, setIsEditDrawerOpen] = useState(false);
  
  // Edit Form state
  const [editName, setEditName] = useState("");
  const [editEmail, setEditEmail] = useState("");
  const [editPhone, setEditPhone] = useState("");

  const { data: customer, isLoading: loadingCustomer } = useQuery({
    queryKey: ["customers", customerId],
    queryFn: async () => {
      const { data, error } = await api.GET("/customers/{id}", {
        params: { path: { id: customerId } }
      });
      if (error) throw error;
      return data as Customer & { subscription?: Subscription };
    },
  });

  const { data: invoices = [], isLoading: loadingInvoices } = useQuery({
    queryKey: ["customers", customerId, "invoices"],
    queryFn: async () => {
      const { data, error } = await (api as any).GET(`/customers/${customerId}/invoices`);
      if (error) throw error;
      return data as Invoice[];
    },
  });

  const { data: plans = [] } = useQuery({
    queryKey: ["plans"],
    queryFn: async () => {
      const { data, error } = await api.GET("/plans");
      if (error) throw error;
      return data as Plan[];
    },
  });

  const generatePortalLink = useMutation({
    mutationFn: async () => {
      const { data, error } = await api.POST("/customers/{id}/portal-link", {
        params: { path: { id: customerId } }
      });
      if (error) throw error;
      return data;
    },
    onSuccess: (data) => {
      if (data?.url) setPortalLink(data.url);
    }
  });

  const updateCustomer = useMutation({
    mutationFn: async () => {
      const { data, error } = await (api as any).PATCH("/customers/{id}", {
        params: { path: { id: customerId } },
        body: { name: editName, email: editEmail, phone: editPhone }
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["customers", customerId] });
      queryClient.invalidateQueries({ queryKey: ["customers"] });
      setIsEditDrawerOpen(false);
    }
  });

  const createSubscription = useMutation({
    mutationFn: async () => {
      const { data, error } = await (api as any).POST("/subscriptions", {
        body: { customer_id: customerId, plan_id: selectedPlanId }
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["customers", customerId] });
      setIsSubDrawerOpen(false);
    }
  });

  const cancelSubscription = useMutation({
    mutationFn: async (subId: string) => {
      const { data, error } = await (api as any).POST(`/subscriptions/${subId}/cancel`, {
        body: { at_period_end: false }
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["customers", customerId] });
    }
  });

  function openEditDrawer() {
    setEditName(customer?.name || "");
    setEditEmail(customer?.email || "");
    setEditPhone(customer?.phone || "");
    setIsEditDrawerOpen(true);
  }

  if (loadingCustomer || !customer) {
    return <div className="p-8 text-center text-slate-500 animate-pulse">Loading customer details...</div>;
  }

  const activeSub = customer.subscription;

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center justify-between mb-4">
          <Link href="/customers" className="inline-flex items-center gap-2 text-sm text-slate-500 hover:text-slate-900 transition-colors">
            <ArrowLeft className="w-4 h-4" /> Back to Customers
          </Link>
          <Button variant="outline" size="sm" onClick={openEditDrawer} className="gap-2">
            <Pencil className="w-4 h-4" /> Edit Profile
          </Button>
        </div>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="h-16 w-16 bg-brand-100 text-brand-600 rounded-full flex items-center justify-center border border-brand-200">
              <User className="h-8 w-8" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-900">{customer.name || customer.email}</h1>
              <p className="text-sm text-slate-500">{customer.email}</p>
              {customer.phone && <p className="text-sm text-slate-500">{customer.phone}</p>}
            </div>
          </div>
          <div className="flex flex-col items-end gap-2">
            <Button 
              variant="outline" 
              onClick={() => generatePortalLink.mutate()}
              disabled={generatePortalLink.isPending}
            >
              {generatePortalLink.isPending ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <LinkIcon className="w-4 h-4 mr-2" />}
              Generate Portal Link
            </Button>
            {portalLink && (
              <a href={portalLink} target="_blank" rel="noopener noreferrer" className="text-xs text-brand-600 hover:underline">
                {portalLink}
              </a>
            )}
          </div>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden flex flex-col">
          <div className="p-6 border-b border-slate-100 flex items-center justify-between">
            <h3 className="font-semibold text-slate-900">Active Subscription</h3>
          </div>
          <div className="p-6 flex-1 flex flex-col items-center justify-center text-center">
            {activeSub && activeSub.status !== "canceled" ? (
              <>
                <div className="w-12 h-12 bg-success-600/10 rounded-full flex items-center justify-center mb-3">
                  <div className="w-3 h-3 bg-success-600 rounded-full"></div>
                </div>
                <h4 className="font-bold text-slate-900">Subscription Active</h4>
                <p className="text-sm text-slate-500 mt-1 capitalize">Status: {activeSub.status}</p>
                <Button 
                  variant="ghost" 
                  size="sm" 
                  className="mt-4 text-danger-600 hover:text-danger-700 hover:bg-danger-50"
                  onClick={() => {
                    if (confirm("Are you sure you want to cancel this subscription immediately?")) {
                      cancelSubscription.mutate(activeSub.id || "");
                    }
                  }}
                  disabled={cancelSubscription.isPending}
                >
                  <Trash2 className="w-4 h-4 mr-2" /> Cancel Subscription
                </Button>
              </>
            ) : (
              <>
                <div className="w-12 h-12 bg-slate-100 rounded-full flex items-center justify-center mb-3">
                  <div className="w-3 h-3 bg-slate-400 rounded-full"></div>
                </div>
                <p className="text-sm text-slate-500 mb-4">No active subscription.</p>
                <Button variant="outline" size="sm" onClick={() => setIsSubDrawerOpen(true)}>
                  <Plus className="w-4 h-4 mr-2" /> Add Subscription
                </Button>
              </>
            )}
          </div>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden flex flex-col">
          <div className="p-6 border-b border-slate-100 flex items-center justify-between">
            <h3 className="font-semibold text-slate-900">Payment Method</h3>
          </div>
          <div className="p-6 flex-1 flex items-center justify-center">
            {customer.has_card_on_file ? (
              <div className="flex items-center gap-3 border p-4 rounded-lg w-full max-w-sm bg-slate-50">
                <CreditCard className="w-6 h-6 text-slate-400" />
                <div>
                  <p className="font-medium text-slate-900">Card on File</p>
                  <p className="text-xs text-slate-500">Ready for automated billing</p>
                </div>
              </div>
            ) : (
              <p className="text-sm text-slate-500 text-center">No card on file. The customer must use the Portal to set up a card or pay via bank transfer.</p>
            )}
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
        <div className="p-6 border-b border-slate-100">
          <h3 className="font-semibold text-slate-900">Invoices</h3>
        </div>
        {loadingInvoices ? (
          <div className="p-8 text-center text-slate-500 animate-pulse">Loading invoices...</div>
        ) : invoices.length === 0 ? (
          <div className="p-12 text-center text-slate-500">
            <p className="text-sm">No invoices found for this customer.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-600">
              <thead className="bg-slate-50 border-b border-slate-200 text-slate-500 font-medium">
                <tr>
                  <th className="px-6 py-4">Amount</th>
                  <th className="px-6 py-4">Status</th>
                  <th className="px-6 py-4">Date</th>
                  <th className="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {invoices.map((inv) => (
                  <tr key={inv.id} className="hover:bg-slate-50 transition-colors">
                    <td className="px-6 py-4 font-medium text-slate-900 tabular-nums">{naira(inv.amount_minor || 0)}</td>
                    <td className="px-6 py-4 capitalize">{inv.status}</td>
                    <td className="px-6 py-4">{formatDate(inv.created_at)}</td>
                    <td className="px-6 py-4 text-right">
                      {inv.invoice_url ? (
                        <a href={inv.invoice_url} target="_blank" rel="noopener noreferrer" className="text-brand-600 hover:underline">
                          View
                        </a>
                      ) : (
                        "—"
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {isSubDrawerOpen && (
        <div className="fixed inset-0 z-50 flex justify-end bg-slate-900/50 backdrop-blur-sm transition-opacity">
          <div className="w-full max-w-md bg-white h-full shadow-2xl flex flex-col animate-in slide-in-from-right duration-300">
            <div className="flex items-center justify-between p-6 border-b border-slate-100">
              <h2 className="text-lg font-bold text-slate-900">Create Subscription</h2>
              <button onClick={() => setIsSubDrawerOpen(false)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-6 flex-1 overflow-y-auto">
              <form id="create-sub-form" onSubmit={(e) => { e.preventDefault(); createSubscription.mutate(); }} className="space-y-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium text-slate-900">Select Plan</label>
                  <select 
                    required
                    className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm ring-offset-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
                    value={selectedPlanId}
                    onChange={(e) => setSelectedPlanId(e.target.value)}
                  >
                    <option value="" disabled>Select a plan...</option>
                    {plans.map(p => (
                      <option key={p.id} value={p.id}>{p.name} ({naira(p.amount_minor || 0)} / {p.interval})</option>
                    ))}
                  </select>
                </div>
              </form>
            </div>
            <div className="p-6 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
              <Button type="button" variant="ghost" onClick={() => setIsSubDrawerOpen(false)}>Cancel</Button>
              <Button type="submit" form="create-sub-form" disabled={createSubscription.isPending || !selectedPlanId}>
                {createSubscription.isPending ? "Creating..." : "Create Subscription"}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Customer Drawer Overlay */}
      {isEditDrawerOpen && (
        <div className="fixed inset-0 z-50 flex justify-end bg-slate-900/50 backdrop-blur-sm transition-opacity">
          <div className="w-full max-w-md bg-white h-full shadow-2xl flex flex-col animate-in slide-in-from-right duration-300">
            <div className="flex items-center justify-between p-6 border-b border-slate-100">
              <h2 className="text-lg font-bold text-slate-900">Edit Customer Profile</h2>
              <button onClick={() => setIsEditDrawerOpen(false)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-6 flex-1 overflow-y-auto">
              <form id="update-customer-form" onSubmit={(e) => { e.preventDefault(); updateCustomer.mutate(); }} className="space-y-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium text-slate-900">Name</label>
                  <Input 
                    placeholder="e.g. John Doe" 
                    value={editName} 
                    onChange={(e) => setEditName(e.target.value)} 
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-slate-900">Email</label>
                  <Input 
                    type="email"
                    required
                    placeholder="john@example.com" 
                    value={editEmail} 
                    onChange={(e) => setEditEmail(e.target.value)} 
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-slate-900">Phone</label>
                  <Input 
                    placeholder="+234..." 
                    value={editPhone} 
                    onChange={(e) => setEditPhone(e.target.value)} 
                  />
                </div>
              </form>
            </div>
            <div className="p-6 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
              <Button type="button" variant="ghost" onClick={() => setIsEditDrawerOpen(false)}>Cancel</Button>
              <Button type="submit" form="update-customer-form" disabled={updateCustomer.isPending}>
                {updateCustomer.isPending ? "Saving..." : "Save Changes"}
              </Button>
            </div>
            {updateCustomer.isError && (
              <div className="p-4 bg-red-50 text-red-700 text-sm">Failed to update customer details.</div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

