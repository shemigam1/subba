"use client";

import { useState } from "react";
import { Key, Plus, Trash2 } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { formatDate } from "@/lib/format";
import type { components } from "@/lib/api/v1";

type APIKey = components["schemas"]["APIKey"];

export default function APIKeysPage() {
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [newKeyData, setNewKeyData] = useState<{key: string} | null>(null);

  const { data: keys = [], isLoading } = useQuery({
    queryKey: ["api-keys"],
    queryFn: async () => {
      const { data, error } = await api.GET("/api-keys");
      if (error) throw error;
      return data;
    },
  });

  const createKey = useMutation({
    mutationFn: async () => {
      const { data, error } = await api.POST("/api-keys", {
        body: { name },
      });
      if (error) throw error;
      return data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      setNewKeyData(data);
      setName("");
    },
  });

  const revokeKey = useMutation({
    mutationFn: async (id: string) => {
      const { error } = await api.DELETE("/api-keys/{id}", {
        params: { path: { id } },
      });
      if (error) throw error;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
    },
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">API Keys</h1>
          <p className="text-sm text-slate-500">
            Manage your secret keys for server-side integration.
          </p>
        </div>
      </div>

      <div className="rounded-xl border border-slate-200 bg-white shadow-sm p-6 space-y-4">
        <h3 className="font-semibold text-slate-900">Generate New Key</h3>
        <form 
          className="flex gap-3 max-w-md"
          onSubmit={(e) => {
            e.preventDefault();
            createKey.mutate();
          }}
        >
          <Input 
            placeholder="e.g. Production Key" 
            required 
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <Button type="submit" disabled={createKey.isPending}>
            {createKey.isPending ? "Generating..." : <><Plus className="w-4 h-4 mr-2" /> Generate</>}
          </Button>
        </form>

        {newKeyData && (
          <div className="mt-4 p-4 rounded-lg bg-warning-600/10 border border-warning-600/20 text-warning-800 space-y-2">
            <h4 className="font-semibold">Key Generated Successfully!</h4>
            <p className="text-sm">Please copy this key immediately. You will not be able to see it again.</p>
            <div className="bg-white p-3 rounded border font-mono text-sm break-all">
              {newKeyData.key}
            </div>
            <Button variant="outline" size="sm" onClick={() => setNewKeyData(null)}>
              I have copied the key
            </Button>
          </div>
        )}
      </div>

      <div className="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center text-slate-500 animate-pulse">Loading keys...</div>
        ) : keys.length === 0 ? (
          <div className="p-12 text-center text-slate-500 flex flex-col items-center">
            <div className="w-12 h-12 bg-slate-100 rounded-full flex items-center justify-center mb-4">
              <Key className="w-6 h-6 text-slate-400" />
            </div>
            <h3 className="font-semibold text-slate-900">No API keys found</h3>
            <p className="text-sm mt-1">Generate a key to authenticate your server requests.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-600">
              <thead className="bg-slate-50 border-b border-slate-200 text-slate-500 font-medium">
                <tr>
                  <th className="px-6 py-4">Name</th>
                  <th className="px-6 py-4">Key Hint</th>
                  <th className="px-6 py-4">Created</th>
                  <th className="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {keys.map((k) => (
                  <tr key={k.id} className="hover:bg-slate-50 transition-colors">
                    <td className="px-6 py-4 font-medium text-slate-900">{k.name}</td>
                    <td className="px-6 py-4 font-mono text-xs">{k.key_hint}</td>
                    <td className="px-6 py-4">{formatDate(k.created_at || "")}</td>
                    <td className="px-6 py-4 text-right">
                      <Button 
                        variant="ghost" 
                        size="sm" 
                        className="text-danger-600 hover:text-danger-700 hover:bg-danger-50"
                        onClick={() => {
                          if (confirm("Are you sure you want to revoke this key? Any integrations using it will fail immediately.")) {
                            revokeKey.mutate(k.id || "");
                          }
                        }}
                      >
                        <Trash2 className="w-4 h-4 mr-2" /> Revoke
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
