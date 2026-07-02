'use client'

import { useState } from 'react'
import { AlertTriangle, Check, Copy, CreditCard, Landmark } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { PortalShell } from '@/components/portal/portal-shell'
import { naira } from '@/lib/format'
import { usePaymentMethod, type ApiError } from '@/lib/portal/hooks'

export default function PaymentMethodPage() {
  return (
    <PortalShell>
      <PaymentMethod />
    </PortalShell>
  )
}

function PaymentMethod() {
  const pm = usePaymentMethod()

  if (pm.isPending) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-40 w-full" />
        <Skeleton className="h-56 w-full" />
      </div>
    )
  }

  if (pm.isError) {
    const err = pm.error as ApiError
    return (
      <Card>
        <CardContent className="pt-6 text-sm">
          <p className="font-medium text-red-700">Couldn&apos;t load your payment methods.</p>
          <p className="mt-1 text-muted-foreground">
            Please try again. {err.requestId && <>Support reference: <code>{err.requestId}</code></>}
          </p>
          <Button variant="outline" className="mt-4" onClick={() => pm.refetch()}>
            Retry
          </Button>
        </CardContent>
      </Card>
    )
  }

  const card = pm.data?.card
  const va = pm.data?.virtual_account

  return (
    <div className="space-y-4">
      {/* Cardless bank transfer — the hero path — comes first, per the design brief. */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Landmark className="h-5 w-5 text-primary" aria-hidden />
            <CardTitle className="text-lg">Pay by bank transfer</CardTitle>
          </div>
          <CardDescription>
            No card needed. Transfer to your personal account below and we&apos;ll detect it
            automatically — your subscription renews on its own.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {va?.account_number ? (
            <>
              <div className="rounded-lg border bg-slate-50 p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">
                  {va.bank_name ?? 'Bank'}
                </p>
                <div className="mt-1 flex items-center justify-between gap-2">
                  <p className="text-2xl font-semibold tabular-nums tracking-wide">
                    {va.account_number}
                  </p>
                  <CopyButton value={va.account_number} label="Copy account number" />
                </div>
                {va.account_name && (
                  <p className="mt-1 text-sm text-muted-foreground">{va.account_name}</p>
                )}
              </div>
              {va.amount_due != null && (va.amount_due.amount_minor ?? 0) > 0 && (
                <p className="text-sm">
                  Transfer <span className="font-semibold tabular-nums">{naira(va.amount_due.amount_minor ?? 0)}</span>{' '}
                  to this account to renew.
                </p>
              )}
            </>
          ) : (
            <p className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
              Your personal account number is being set up. Check back shortly — it will appear
              here as soon as it&apos;s ready.
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <CreditCard className="h-5 w-5 text-primary" aria-hidden />
            <CardTitle className="text-lg">Card on file</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {card ? (
            <div className="flex items-center justify-between rounded-lg border p-4">
              <div>
                <p className="font-medium capitalize">
                  {card.brand} •••• {card.last4}
                </p>
                {card.exp_month != null && card.exp_year != null && card.exp_month > 0 && (
                  <p className="mt-0.5 text-sm text-muted-foreground">
                    Expires {String(card.exp_month).padStart(2, '0')}/{card.exp_year}
                  </p>
                )}
              </div>
              {card.expired && (
                <span className="inline-flex items-center gap-1.5 rounded-full border border-amber-200 bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700">
                  <AlertTriangle className="h-3.5 w-3.5" aria-hidden />
                  Expired
                </span>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No card on file.</p>
          )}
          {/* Nomba's hosted tokenization widget mounts here once Track A's client lands. */}
          <Button variant="outline" className="h-11 w-full" disabled title="Card updates are coming soon">
            {card ? 'Update card' : 'Add a card'} (coming soon)
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}

function CopyButton({ value, label }: { value: string; label: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <button
      aria-label={label}
      onClick={async () => {
        await navigator.clipboard.writeText(value)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      }}
      className="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-lg border bg-background text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      {copied ? <Check className="h-4 w-4 text-green-600" aria-hidden /> : <Copy className="h-4 w-4" aria-hidden />}
    </button>
  )
}
