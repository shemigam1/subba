"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowLeft, User, CreditCard, Link as LinkIcon, Loader2 } from "lucide-react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
import type { components } from "@/lib/api/v1";

type Customer = components["schemas"]["Customer"];
type Subscription = components["schemas"]["Subscription"];

export default function CustomerDetailPage({ params }: { params: { id: string } }) {
  const customerId = params.id;
  const [portalLink, setPortalLink] = useState("");

  const { data: customer, isLoading: loadingCustomer } = useQuery({
    queryKey: ["customers", customerId],
    queryFn: async () => {
      const { data, error } = await api.GET("/customers/{id}", {
        params: { path: { id: customerId } }
      });
      if (error) throw error;
      return data as Customer;
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

  if (loadingCustomer || !customer) {
    return <div className="p-8 text-center text-slate-500 animate-pulse">Loading customer details...</div>;
  }

  return (
    <div className="space-y-6">
      <div>
        <Link href="/customers" className="inline-flex items-center gap-2 text-sm text-slate-500 hover:text-slate-900 mb-4 transition-colors">
          <ArrowLeft className="w-4 h-4" /> Back to Customers
        </Link>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="h-16 w-16 bg-brand-100 text-brand-600 rounded-full flex items-center justify-center border border-brand-200">
              <User className="h-8 w-8" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-900">{customer.name || customer.email}</h1>
              <p className="text-sm text-slate-500">{customer.email}</p>
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
            {/* Minimal placeholder for subscription details, since we haven't mocked a complex subscription endpoint yet */}
            <div className="w-12 h-12 bg-success-600/10 rounded-full flex items-center justify-center mb-3">
              <div className="w-3 h-3 bg-success-600 rounded-full"></div>
            </div>
            <h4 className="font-bold text-slate-900">Pro Plan</h4>
            <p className="text-sm text-slate-500 mt-1">Status: Active</p>
          </div>
        </div>

        <div className="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden flex flex-col">
          <div className="p-6 border-b border-slate-100 flex items-center justify-between">
            <h3 className="font-semibold text-slate-900">Payment Method</h3>
          </div>
          <div className="p-6 flex-1 flex items-center justify-center">
            {customer.has_card_on_file ? (
              <div className="flex items-center gap-3 border p-4 rounded-lg w-full max-w-sm">
                <CreditCard className="w-6 h-6 text-slate-400" />
                <div>
                  <p className="font-medium text-slate-900">Card on File</p>
                  <p className="text-xs text-slate-500">Ready for automated billing</p>
                </div>
              </div>
            ) : (
              <p className="text-sm text-slate-500 text-center">No card on file. Using Bank Transfer.</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

