'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { useEffect, type ReactNode } from 'react'
import { CreditCard, LogOut, Receipt, Wallet } from 'lucide-react'

import { Skeleton } from '@/components/ui/skeleton'
import { usePortalMe, usePortalLogout } from '@/lib/portal/hooks'
import { cn } from '@/lib/utils'

const nav = [
  { href: '/pay', label: 'Subscription', icon: Wallet },
  { href: '/pay/invoices', label: 'Invoices', icon: Receipt },
  { href: '/pay/payment-method', label: 'Payment', icon: CreditCard },
]

// PortalShell guards every signed-in portal screen: it loads the session context
// (customer + tenant branding), bounces to /pay/access when there is none, and
// renders the co-branded chrome (tenant's name, not Subba's).
export function PortalShell({ children }: { children: ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const me = usePortalMe()
  const logout = usePortalLogout()

  const unauthenticated = me.isError && (me.error as { status?: number }).status === 401
  useEffect(() => {
    if (unauthenticated) router.replace('/pay/access')
  }, [unauthenticated, router])

  if (me.isPending || unauthenticated) {
    return (
      <div className="mx-auto w-full max-w-xl px-4 py-6">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="mt-6 h-48 w-full" />
      </div>
    )
  }

  const tenantName = me.data?.tenant_branding?.tenant_name ?? 'Your provider'

  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-xl flex-col px-4">
      <header className="flex items-center justify-between py-5">
        <div>
          <p className="text-lg font-semibold tracking-tight">{tenantName}</p>
          <p className="text-xs text-muted-foreground">{me.data?.customer?.email}</p>
        </div>
        <button
          onClick={() => logout.mutate(undefined, { onSuccess: () => router.replace('/pay/access') })}
          className="inline-flex h-11 items-center gap-1.5 rounded-lg px-3 text-sm text-muted-foreground hover:bg-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <LogOut className="h-4 w-4" aria-hidden />
          Sign out
        </button>
      </header>

      <nav aria-label="Portal" className="mb-6 grid grid-cols-3 gap-1 rounded-xl border bg-card p-1">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || (href !== '/pay' && pathname.startsWith(href))
          return (
            <Link
              key={href}
              href={href}
              aria-current={active ? 'page' : undefined}
              className={cn(
                'inline-flex h-11 items-center justify-center gap-1.5 rounded-lg text-sm font-medium transition-colors',
                active
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-accent hover:text-foreground'
              )}
            >
              <Icon className="h-4 w-4" aria-hidden />
              {label}
            </Link>
          )
        })}
      </nav>

      <main className="flex-1 pb-10">{children}</main>
    </div>
  )
}
