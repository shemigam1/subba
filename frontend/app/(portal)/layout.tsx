import { ReactNode } from "react";

export default function PortalLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-slate-50 flex flex-col font-sans text-slate-900">
      {/* Minimal Top Header - Co-branded */}
      <header className="h-16 flex items-center justify-between px-4 sm:px-6 max-w-3xl w-full mx-auto border-b border-transparent">
        <div className="flex items-center gap-2">
          {/* Tenant Logo Placeholder */}
          <div className="w-8 h-8 bg-slate-900 rounded-md flex items-center justify-center text-white font-bold text-sm">
            Ac
          </div>
          <span className="font-semibold text-lg tracking-tight">Acme Corp</span>
        </div>
        <div className="text-xs text-slate-500 flex items-center gap-1">
          Powered by <span className="font-semibold text-slate-900">Subba</span>
        </div>
      </header>

      {/* Main Content Area - Mobile First */}
      <main className="flex-1 flex flex-col items-center justify-start p-4 sm:p-6 sm:mt-8">
        <div className="w-full max-w-md">
          {children}
        </div>
      </main>
    </div>
  );
}
