'use client'

import Link from 'next/link'
import { ChevronRight, ReceiptText } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { PortalShell } from '@/components/portal/portal-shell'
import { StatusBadge } from '@/components/portal/status-badge'
import { formatDate, naira } from '@/lib/format'
import { useInvoices, type ApiError } from '@/lib/portal/hooks'

export default function InvoicesPage() {
  return (
    <PortalShell>
      <Invoices />
    </PortalShell>
  )
}

function Invoices() {
  const invoices = useInvoices()

  if (invoices.isPending) {
    return (
      <div className="space-y-3">
        {[0, 1, 2].map((i) => (
          <Skeleton key={i} className="h-20 w-full" />
        ))}
      </div>
    )
  }

  if (invoices.isError) {
    const err = invoices.error as ApiError
    return (
      <Card>
        <CardContent className="pt-6 text-sm">
          <p className="font-medium text-red-700">Couldn&apos;t load your invoices.</p>
          <p className="mt-1 text-muted-foreground">
            Please try again. {err.requestId && <>Support reference: <code>{err.requestId}</code></>}
          </p>
          <Button variant="outline" className="mt-4" onClick={() => invoices.refetch()}>
            Retry
          </Button>
        </CardContent>
      </Card>
    )
  }

  if (invoices.data.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center py-12 text-center">
          <ReceiptText className="h-10 w-10 text-muted-foreground" aria-hidden />
          <p className="mt-4 font-medium">No invoices yet</p>
          <p className="mt-1 max-w-xs text-sm text-muted-foreground">
            Your payment history will appear here after your first billing cycle.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <ul className="space-y-3">
      {invoices.data.map((inv) => (
        <li key={inv.id}>
          <Link
            href={`/pay/invoices/${inv.id}`}
            className="flex items-center justify-between gap-3 rounded-xl border bg-card p-4 transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <div>
              <p className="font-medium tabular-nums">{naira(inv.amount_minor ?? 0)}</p>
              <p className="mt-0.5 text-sm text-muted-foreground">
                {inv.issued_at ? formatDate(inv.issued_at) : '—'}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <StatusBadge status={inv.status} />
              <ChevronRight className="h-4 w-4 text-muted-foreground" aria-hidden />
            </div>
          </Link>
        </li>
      ))}
    </ul>
  )
}
