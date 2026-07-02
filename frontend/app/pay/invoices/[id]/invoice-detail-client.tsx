'use client'

import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { PortalShell } from '@/components/portal/portal-shell'
import { StatusBadge } from '@/components/portal/status-badge'
import { formatDate, naira } from '@/lib/format'
import { useInvoice, type ApiError } from '@/lib/portal/hooks'

export function InvoiceDetailClient({ id }: { id: string }) {
  return (
    <PortalShell>
      <InvoiceDetail id={id} />
    </PortalShell>
  )
}

function InvoiceDetail({ id }: { id: string }) {
  const invoice = useInvoice(id)

  if (invoice.isPending) {
    return <Skeleton className="h-64 w-full" />
  }

  if (invoice.isError) {
    const err = invoice.error as ApiError
    return (
      <Card>
        <CardContent className="pt-6 text-sm">
          <p className="font-medium text-red-700">
            {err.status === 404 ? 'Invoice not found.' : "Couldn't load this invoice."}
          </p>
          <p className="mt-1 text-muted-foreground">
            {err.requestId && <>Support reference: <code>{err.requestId}</code></>}
          </p>
          <Button asChild variant="outline" className="mt-4">
            <Link href="/pay/invoices">
              <ArrowLeft className="h-4 w-4" aria-hidden />
              Back to invoices
            </Link>
          </Button>
        </CardContent>
      </Card>
    )
  }

  const inv = invoice.data
  if (!inv) return null

  return (
    <div className="space-y-4">
      <Link
        href="/pay/invoices"
        className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" aria-hidden />
        All invoices
      </Link>

      <Card>
        <CardHeader className="flex-row items-start justify-between space-y-0">
          <div>
            <p className="text-sm text-muted-foreground">
              {inv.issued_at ? formatDate(inv.issued_at) : 'Invoice'}
            </p>
            <p className="mt-1 text-3xl font-semibold tabular-nums tracking-tight">
              {naira(inv.amount_minor ?? 0)}
            </p>
            {inv.period_start && inv.period_end && (
              <p className="mt-1 text-sm text-muted-foreground">
                {formatDate(inv.period_start)} – {formatDate(inv.period_end)}
              </p>
            )}
          </div>
          <StatusBadge status={inv.status} />
        </CardHeader>
        {inv.items && inv.items.length > 0 && (
          <CardContent>
            <p className="mb-2 text-sm font-medium">Details</p>
            <ul className="divide-y rounded-lg border">
              {inv.items.map((item) => (
                <li key={item.id} className="flex items-start justify-between gap-3 p-3 text-sm">
                  <div>
                    <p>{item.description}</p>
                    {item.period_start && item.period_end && (
                      <p className="mt-0.5 text-xs text-muted-foreground">
                        {formatDate(item.period_start)} – {formatDate(item.period_end)}
                      </p>
                    )}
                  </div>
                  {/* Negative amounts are proration credits — keep the minus visible. */}
                  <p className="shrink-0 tabular-nums">{naira(item.amount_minor ?? 0)}</p>
                </li>
              ))}
            </ul>
          </CardContent>
        )}
      </Card>
    </div>
  )
}
