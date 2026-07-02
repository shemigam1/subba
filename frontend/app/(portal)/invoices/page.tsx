"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowLeft, ExternalLink, Download, Clock, CheckCircle2, XCircle } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { naira } from "@/lib/format/money";
import type { components } from "@/lib/api/types";

type Invoice = components["schemas"]["Invoice"];

export default function InvoicesPage() {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchInvoices() {
      const { data, error } = await apiClient.GET("/portal/invoices");
      if (data) {
        setInvoices(data);
      } else {
        console.error(error);
      }
      setLoading(false);
    }
    fetchInvoices();
  }, []);

  if (loading) {
    return <div className="p-8 text-center text-slate-500 animate-pulse">Loading invoices...</div>;
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div>
        <Link href="/pay" className="inline-flex items-center gap-2 text-sm text-slate-500 hover:text-slate-900 mb-4 transition-colors">
          <ArrowLeft className="w-4 h-4" /> Back to Subscription
        </Link>
        <h1 className="text-2xl font-bold tracking-tight text-slate-900">Invoice History</h1>
        <p className="text-sm text-slate-500">View your past payments and download receipts.</p>
      </div>

      <div className="space-y-4">
        {invoices.length === 0 ? (
          <div className="rounded-xl border border-slate-200 bg-white p-8 text-center text-slate-500 shadow-sm">
            You don't have any invoices yet.
          </div>
        ) : (
          <div className="hidden md:block rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
            <table className="w-full text-left text-sm text-slate-600">
              <thead className="bg-slate-50 border-b border-slate-200 text-slate-500 font-medium">
                <tr>
                  <th className="px-6 py-4">Date</th>
                  <th className="px-6 py-4">Amount</th>
                  <th className="px-6 py-4">Status</th>
                  <th className="px-6 py-4 text-right">Receipt</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {invoices.map((inv) => (
                  <tr key={inv.id} className="hover:bg-slate-50 transition-colors">
                    <td className="px-6 py-4 font-medium text-slate-900">
                      {inv.issued_at ? new Date(inv.issued_at).toLocaleDateString("en-NG", { year: 'numeric', month: 'short', day: 'numeric' }) : "Unknown Date"}
                    </td>
                    <td className="px-6 py-4 tabular-nums">{naira(inv.amount_minor || 0)}</td>
                    <td className="px-6 py-4">
                      {inv.status === "paid" ? (
                        <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium bg-success-600/10 text-success-600">
                          <CheckCircle2 className="w-3 h-3" /> Paid
                        </span>
                      ) : inv.status === "open" ? (
                        <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium bg-info-600/10 text-info-600">
                          <Clock className="w-3 h-3" /> Due
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium bg-slate-100 text-slate-600">
                          <XCircle className="w-3 h-3" /> {inv.status}
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <button className="inline-flex items-center gap-1 text-brand-600 hover:text-brand-700 font-medium text-xs">
                        <Download className="w-3 h-3" /> PDF
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Mobile View: Stacked Cards */}
        <div className="md:hidden space-y-3">
          {invoices.map((inv) => (
            <div key={inv.id} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm flex items-center justify-between">
              <div>
                <div className="font-medium text-slate-900">
                  {inv.issued_at ? new Date(inv.issued_at).toLocaleDateString("en-NG", { year: 'numeric', month: 'short', day: 'numeric' }) : "Unknown Date"}
                </div>
                <div className="text-sm font-bold mt-1 tabular-nums">
                  {naira(inv.amount_minor || 0)}
                </div>
                <div className="mt-2">
                  {inv.status === "paid" ? (
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium bg-success-600/10 text-success-600">
                      <CheckCircle2 className="w-3 h-3" /> Paid
                    </span>
                  ) : inv.status === "open" ? (
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium bg-info-600/10 text-info-600">
                      <Clock className="w-3 h-3" /> Due
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium bg-slate-100 text-slate-600">
                      <XCircle className="w-3 h-3" /> {inv.status}
                    </span>
                  )}
                </div>
              </div>
              <button className="p-2 text-slate-400 hover:text-brand-600 bg-slate-50 rounded-full transition-colors">
                <Download className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
