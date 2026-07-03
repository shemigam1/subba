'use client'

import Link from 'next/link'
import { useState } from 'react'
import { AlertTriangle, CreditCard, Loader2, PackageOpen } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { PortalShell } from '@/components/portal/portal-shell'
import { StatusBadge } from '@/components/portal/status-badge'
import { formatDate, intervalLabel, naira } from '@/lib/format'
import { useCancelSubscription, useSubscription, type ApiError } from '@/lib/portal/hooks'

export default function SubscriptionHomePage() {
  return (
    <PortalShell>
      <SubscriptionHome />
    </PortalShell>
  )
}

function SubscriptionHome() {
  const sub = useSubscription()

  if (sub.isPending) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-56 w-full" />
        <Skeleton className="h-12 w-full" />
      </div>
    )
  }

  if (sub.isError) {
    const err = sub.error as ApiError
    return (
      <Card>
        <CardContent className="pt-6 text-sm">
          <p className="font-medium text-red-700">Couldn&apos;t load your subscription.</p>
          <p className="mt-1 text-muted-foreground">
            Please try again. {err.requestId && <>Support reference: <code>{err.requestId}</code></>}
          </p>
          <Button variant="outline" className="mt-4" onClick={() => sub.refetch()}>
            Retry
          </Button>
        </CardContent>
      </Card>
    )
  }

  if (!sub.data) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center py-12 text-center">
          <PackageOpen className="h-10 w-10 text-muted-foreground" aria-hidden />
          <p className="mt-4 font-medium">No subscription yet</p>
          <p className="mt-1 max-w-xs text-sm text-muted-foreground">
            When your provider starts a subscription for you, it will appear here.
          </p>
        </CardContent>
      </Card>
    )
  }

  const s = sub.data
  const pastDue = s.status === 'past_due'
  const canceled = s.status === 'canceled' || s.cancel_at_period_end

  return (
    <div className="space-y-4">
      {pastDue && (
        <div className="flex items-start gap-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden />
          <div>
            <p className="font-medium">Your last payment didn&apos;t go through</p>
            <p className="mt-1">
              Renew now by bank transfer — no card needed.{' '}
              <Link href="/pay/payment-method" className="font-medium underline underline-offset-2">
                Pay by bank transfer
              </Link>
            </p>
          </div>
        </div>
      )}

      <Card>
        <CardHeader className="flex-row items-start justify-between space-y-0">
          <div>
            <p className="text-sm text-muted-foreground">{s.plan?.name ?? 'Your plan'}</p>
            <p className="mt-1 text-3xl font-semibold tabular-nums tracking-tight">
              {naira(s.plan?.amount_minor ?? 0)}
              <span className="text-base font-normal text-muted-foreground">
                {intervalLabel(s.plan?.interval)}
              </span>
            </p>
          </div>
          <StatusBadge status={s.status} />
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            {canceled && s.current_period_end
              ? `Active until ${formatDate(s.current_period_end)} — then your subscription ends.`
              : s.current_period_end
                ? `Renews ${formatDate(s.current_period_end)}`
                : 'Renewal date pending'}
          </p>
          <div className="flex flex-col gap-2 sm:flex-row">
            <Button asChild className="h-11 flex-1">
              <Link href="/pay/payment-method">
                <CreditCard className="h-4 w-4" aria-hidden />
                Update payment method
              </Link>
            </Button>
            {!canceled && <CancelDialog periodEnd={s.current_period_end} />}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function CancelDialog({ periodEnd }: { periodEnd?: string | null }) {
  const [open, setOpen] = useState(false)
  const cancel = useCancelSubscription()

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <Button variant="outline" className="h-11 flex-1" onClick={() => setOpen(true)}>
        Cancel subscription
      </Button>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Cancel your subscription?</DialogTitle>
          <DialogDescription>
            {periodEnd
              ? `You'll keep full access until ${formatDate(periodEnd)}. After that, no further charges.`
              : "You'll keep access until the end of the period you've paid for."}
          </DialogDescription>
        </DialogHeader>
        {cancel.isError && (
          <p className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
            Couldn&apos;t cancel — please try again.
            {(cancel.error as ApiError).requestId && (
              <> Support reference: <code>{(cancel.error as ApiError).requestId}</code></>
            )}
          </p>
        )}
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>
            Keep my subscription
          </Button>
          <Button
            variant="destructive"
            disabled={cancel.isPending}
            onClick={() => cancel.mutate(true, { onSuccess: () => setOpen(false) })}
          >
            {cancel.isPending && <Loader2 className="h-4 w-4 animate-spin" aria-hidden />}
            Cancel at period end
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
