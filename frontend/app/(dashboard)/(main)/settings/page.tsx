"use client";

import { useState, useEffect } from "react";
import { Save, Loader2, Link as LinkIcon, Settings2 } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
interface Settings {
  business_name?: string;
  support_email?: string;
  webhook_url?: string;
}

export default function SettingsPage() {
  const queryClient = useQueryClient();
  const [businessName, setBusinessName] = useState("");
  const [supportEmail, setSupportEmail] = useState("");
  const [webhookUrl, setWebhookUrl] = useState("");

  const { data: settings, isLoading } = useQuery({
    queryKey: ["settings"],
    queryFn: async () => {
      const { data, error } = await api.GET("/settings");
      if (error) throw error;
      return data as Settings;
    },
  });

  useEffect(() => {
    if (settings) {
      setBusinessName(settings.business_name || "");
      setSupportEmail(settings.support_email || "");
      setWebhookUrl(settings.webhook_url || "");
    }
  }, [settings]);

  const updateSettings = useMutation({
    mutationFn: async () => {
      const { data, error } = await api.PATCH("/settings", {
        body: {
          business_name: businessName,
          support_email: supportEmail,
          webhook_url: webhookUrl,
        }
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["settings"] });
    },
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    updateSettings.mutate();
  }

  if (isLoading) {
    return <div className="p-8 text-center text-slate-500 animate-pulse">Loading settings...</div>;
  }

  return (
    <div className="space-y-6 max-w-3xl">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-900">Settings</h1>
        <p className="text-sm text-slate-500">
          Manage your business profile and integrations.
        </p>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
        <div className="p-6 border-b border-slate-100 flex items-center gap-3">
          <Settings2 className="w-5 h-5 text-slate-400" />
          <h3 className="font-semibold text-slate-900">General Settings</h3>
        </div>
        <div className="p-6">
          <form id="settings-form" onSubmit={handleSubmit} className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-900">Business Name</label>
                <Input 
                  required 
                  value={businessName} 
                  onChange={(e) => setBusinessName(e.target.value)} 
                  placeholder="Acme Corp"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-slate-900">Support Email</label>
                <Input 
                  required 
                  type="email"
                  value={supportEmail} 
                  onChange={(e) => setSupportEmail(e.target.value)} 
                  placeholder="support@example.com"
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-900 flex items-center gap-2">
                <LinkIcon className="w-4 h-4 text-slate-400" /> Webhook URL
              </label>
              <Input 
                type="url"
                value={webhookUrl} 
                onChange={(e) => setWebhookUrl(e.target.value)} 
                placeholder="https://your-domain.com/webhooks/subba"
              />
              <p className="text-xs text-slate-500">
                We will send POST requests to this URL whenever a subscription event occurs.
              </p>
            </div>
          </form>
        </div>
        <div className="p-6 border-t border-slate-100 bg-slate-50 flex justify-end items-center gap-4">
          {updateSettings.isSuccess && (
            <span className="text-sm text-success-600 font-medium">Settings saved!</span>
          )}
          {updateSettings.isError && (
            <span className="text-sm text-danger-600 font-medium">Failed to save settings.</span>
          )}
          <Button type="submit" form="settings-form" disabled={updateSettings.isPending}>
            {updateSettings.isPending ? (
              <><Loader2 className="w-4 h-4 mr-2 animate-spin" /> Saving...</>
            ) : (
              <><Save className="w-4 h-4 mr-2" /> Save Changes</>
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
