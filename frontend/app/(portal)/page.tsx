"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { CheckCircle2, CreditCard, ExternalLink, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/api/client";
import { naira } from "@/lib/format/money";
import type { components } from "@/lib/api/types";

type Subscription = components["schemas"]["Subscription"];

export default function PortalHomePage() {
  const [sub, setSub] = useState<Subscription | null>(null);
  const [loading, setLoading] = useState(true);
  const [isCancelModalOpen, setIsCancelModalOpen] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);

  useEffect(() => {
    async function fetchSub() {
      const { data, error } = await apiClient.GET("/portal/subscription");
      if (data) {
        setSub(data);
      } else {
        console.error(error);
      }
      setLoading(false);
    }
    fetchSub();
  }, []);

  async function handleCancel() {
    setIsCanceling(true);
    const { data, error } = await apiClient.POST("/portal/subscription/cancel", {
      body: { at_period_end: true }
    });
    if (data) {
      setSub(data);
      setIsCancelModalOpen(false);
    } else {
      console.error(error);
    }
    setIsCanceling(false);
  }

  if (loading || !sub) {
    return <div className="p-8 text-center text-slate-500 animate-pulse">Loading subscription details...</div>;
  }

  const isCanceled = sub.status === "canceled" || sub.cancel_at_period_end;
  const plan = sub.plan;

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      {/* Subscription Hero Card */}
      <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden relative">
        {isCanceled && (
          <div className="bg-warning-600 text-white px-6 py-2 text-sm font-medium text-center">
            Your subscription will be canceled at the end of the billing period.
          </div>
        )}
        <div className="p-6 sm:p-8">
          <div className="flex items-start justify-between mb-6">
            <div>
              <p className="text-sm font-medium text-slate-500 mb-1">Current Plan</p>
              <h2 className="text-2xl font-bold text-slate-900 tracking-tight">{plan?.name || "Unknown Plan"}</h2>
            </div>
            
            {/* Status Badge */}
            {isCanceled ? (
              <div className="flex items-center gap-1.5 bg-slate-100 text-slate-600 px-3 py-1 rounded-full text-sm font-medium">
                Canceled
              </div>
            ) : (
              <div className="flex items-center gap-1.5 bg-success-600/10 text-success-600 px-3 py-1 rounded-full text-sm font-medium">
                <CheckCircle2 className="w-4 h-4" />
                Active
              </div>
            )}
          </div>

          <div className="mb-8">
            <div className="flex items-baseline gap-1">
              <span className="text-4xl font-bold text-slate-900 tabular-nums">
                {plan ? naira(plan.amount_minor || 0) : "₦0.00"}
              </span>
              <span className="text-slate-500 font-medium">/{plan?.interval === 'year' ? 'yr' : 'mo'}</span>
            </div>
            {!isCanceled && sub.current_period_end && (
              <p className="text-slate-500 mt-2 font-medium">
                Renews <span className="text-slate-900">
                  {new Date(sub.current_period_end).toLocaleDateString('en-NG', { day: 'numeric', month: 'short', year: 'numeric' })}
                </span>
              </p>
            )}
            {isCanceled && sub.current_period_end && (
              <p className="text-slate-500 mt-2 font-medium">
                Active until <span className="text-slate-900">
                  {new Date(sub.current_period_end).toLocaleDateString('en-NG', { day: 'numeric', month: 'short', year: 'numeric' })}
                </span>
              </p>
            )}
          </div>

          <div className="space-y-3">
            <Button className="w-full h-12 text-base shadow-sm">
              <CreditCard className="w-4 h-4 mr-2" />
              Update payment method
            </Button>
            {!isCanceled && (
              <Button 
                variant="ghost" 
                className="w-full h-12 text-slate-500 hover:text-danger-600 hover:bg-danger-50 transition-colors"
                onClick={() => setIsCancelModalOpen(true)}
              >
                Cancel subscription
              </Button>
            )}
            {isCanceled && (
              <Button 
                variant="outline" 
                className="w-full h-12 text-brand-600 border-brand-200 hover:bg-brand-50 transition-colors"
              >
                Resubscribe
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Quick Links / Secondary Actions */}
      <div className="grid grid-cols-2 gap-4">
        <Link href="/invoices" className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm flex flex-col items-start gap-2 hover:border-slate-300 hover:bg-slate-50 transition-colors">
          <div className="w-8 h-8 rounded-full bg-slate-100 text-slate-600 flex items-center justify-center mb-2">
            <ExternalLink className="w-4 h-4" />
          </div>
          <span className="font-medium text-slate-900 text-sm">Billing history</span>
          <span className="text-xs text-slate-500 text-left">View past invoices and receipts</span>
        </Link>
        <button className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm flex flex-col items-start gap-2 hover:border-slate-300 hover:bg-slate-50 transition-colors">
          <div className="w-8 h-8 rounded-full bg-slate-100 text-slate-600 flex items-center justify-center mb-2">
            <CreditCard className="w-4 h-4" />
          </div>
          <span className="font-medium text-slate-900 text-sm">Payment methods</span>
          <span className="text-xs text-slate-500 text-left">Manage cards and virtual accounts</span>
        </button>
      </div>

      {/* Cancel Modal Overlay */}
      {isCancelModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm transition-opacity">
          <div className="w-full max-w-sm bg-white rounded-2xl shadow-2xl p-6 animate-in zoom-in-95 duration-200">
            <div className="w-12 h-12 rounded-full bg-warning-100 text-warning-600 flex items-center justify-center mb-4">
              <AlertTriangle className="w-6 h-6" />
            </div>
            <h3 className="text-xl font-bold text-slate-900 mb-2">Cancel Subscription?</h3>
            <p className="text-slate-500 mb-6 text-sm leading-relaxed">
              You will continue to have full access to your plan until the end of your current billing period on 
              {" "}<span className="font-semibold text-slate-700">{sub.current_period_end ? new Date(sub.current_period_end).toLocaleDateString('en-NG', { day: 'numeric', month: 'short', year: 'numeric' }) : "the end of the period"}</span>. 
              After that, your account will be downgraded.
            </p>
            <div className="flex flex-col gap-3">
              <Button 
                variant="destructive" 
                className="w-full h-11" 
                onClick={handleCancel}
                disabled={isCanceling}
              >
                {isCanceling ? "Canceling..." : "Yes, cancel subscription"}
              </Button>
              <Button 
                variant="ghost" 
                className="w-full h-11 text-slate-600" 
                onClick={() => setIsCancelModalOpen(false)}
                disabled={isCanceling}
              >
                Keep my subscription
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
