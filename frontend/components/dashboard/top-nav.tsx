"use client";

import { Bell, Search } from "lucide-react";
import { Input } from "@/components/ui/input";

export function TopNav() {
  return (
    <header className="h-16 flex items-center justify-between px-6 border-b border-slate-200 bg-white">
      <div className="flex-1 max-w-md flex items-center">
        {/* We can use Shadcn Input when it's fully ready, for now simple HTML input */}
        <div className="relative w-full">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-slate-500" />
          <input
            type="text"
            placeholder="Search customers, plans..."
            className="w-full bg-slate-50 border border-slate-200 rounded-md pl-9 pr-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all"
          />
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 bg-slate-100 rounded-full p-1 border border-slate-200">
          <button className="px-3 py-1 text-xs font-medium bg-white shadow-sm rounded-full text-slate-900">
            Live
          </button>
          <button className="px-3 py-1 text-xs font-medium text-slate-500 hover:text-slate-900 rounded-full transition-colors">
            Test
          </button>
        </div>
        <button className="text-slate-500 hover:text-slate-900 relative">
          <Bell className="w-5 h-5" />
          <span className="absolute top-0 right-0 w-2 h-2 bg-red-500 rounded-full border border-white"></span>
        </button>
      </div>
    </header>
  );
}
