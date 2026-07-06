"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useUser } from "@/lib/hooks/use-user";

// Client-side guard for the dashboard: /me is called with the Bearer tenant token;
// if it fails (no/expired token), redirect to /login. Rendered inside the (server)
// dashboard layout so the pages it wraps stay server components.
export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { data, isLoading, isError } = useUser();

  useEffect(() => {
    if (isError) router.replace("/login");
  }, [isError, router]);

  if (isLoading) {
    return (
      <div className="flex h-full w-full items-center justify-center py-24 text-sm text-slate-500">
        Loading…
      </div>
    );
  }

  if (isError || !data) return null;

  return <>{children}</>;
}
