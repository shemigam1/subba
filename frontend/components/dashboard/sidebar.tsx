"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  Users,
  CreditCard,
  Key,
  Settings,
  LogOut,
} from "lucide-react";
import { api } from "@/lib/api";
import { clearTenantToken } from "@/lib/auth/token";
import { useUser } from "@/lib/hooks/use-user";

// In a real shadcn setup with itshover, these would be animated ItsHover icons.
const navItems = [
  { name: "Overview", href: "/overview", icon: LayoutDashboard },
  { name: "Plans", href: "/plans", icon: CreditCard },
  { name: "Customers", href: "/customers", icon: Users },
  { name: "API Keys", href: "/api-keys", icon: Key },
  { name: "Settings", href: "/settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { data: user } = useUser();

  async function signOut() {
    // Best-effort server-side revoke; the local token clear is what matters.
    try {
      await api.POST("/auth/logout");
    } catch {
      /* ignore — token clear below still signs us out locally */
    }
    clearTenantToken();
    router.replace("/login");
  }

  const initials = (user?.name ?? "?")
    .split(/\s+/)
    .map((w) => w[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();

  return (
    <aside className="w-64 flex-shrink-0 border-r border-slate-200 bg-slate-50 flex flex-col h-screen">
      <div className="h-16 flex items-center px-6 border-b border-slate-200">
        {/* Placeholder for Subba Logo */}
        <div className="font-bold text-xl tracking-tight text-slate-900 flex items-center gap-2">
          <div className="w-6 h-6 bg-primary rounded-sm" />
          Subba
        </div>
      </div>

      <nav className="flex-1 overflow-y-auto py-4 px-3 space-y-1">
        {navItems.map((item) => {
          const isActive = pathname.startsWith(item.href);
          const Icon = item.icon;
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors",
                isActive
                  ? "bg-primary/10 text-primary"
                  : "text-slate-700 hover:bg-slate-100 hover:text-slate-900"
              )}
            >
              <Icon className={cn("w-5 h-5", isActive ? "text-primary" : "text-slate-500")} />
              {item.name}
            </Link>
          );
        })}
      </nav>

      <div className="p-4 border-t border-slate-200">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-slate-200 flex items-center justify-center text-slate-600 font-medium text-xs">
            {initials}
          </div>
          <div className="flex flex-col min-w-0">
            <span className="text-sm font-medium text-slate-900 truncate">{user?.name ?? "…"}</span>
            <span className="text-xs text-slate-500 truncate">{user?.email ?? ""}</span>
          </div>
          <button
            onClick={signOut}
            title="Sign out"
            className="ml-auto text-slate-500 hover:text-slate-900 transition-colors"
          >
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </div>
    </aside>
  );
}
