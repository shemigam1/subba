import type { Metadata } from 'next'
import type { ReactNode } from 'react'

export const metadata: Metadata = {
  title: 'Manage your subscription',
  description: 'Securely view and manage your subscription.',
}

// Mobile-first portal canvas: a calm slate background with a centered column.
// The signed-in chrome lives in PortalShell (client) so the access page stays bare.
export default function PortalLayout({ children }: { children: ReactNode }) {
  return <div className="min-h-dvh bg-slate-50 text-base">{children}</div>
}
