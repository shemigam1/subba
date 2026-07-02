'use client'

import { useRouter } from 'next/navigation'
import { useEffect, useState, useSyncExternalStore, type FormEvent } from 'react'
import { Loader2, MailCheck, ShieldCheck } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useExchangeToken, useRequestAccess } from '@/lib/portal/hooks'

// The tenant id travels on the entry link (?t=...); remember it so a customer
// whose link expired can request a fresh one without the original URL.
const TENANT_KEY = 'subba_portal_tenant'

export function AccessClient({ token, tenantId }: { token?: string; tenantId?: string }) {
  const router = useRouter()
  const exchange = useExchangeToken()
  const request = useRequestAccess()
  const [email, setEmail] = useState('')

  // Fall back to the tenant id remembered from a previous visit; reads null on the
  // server so SSR and hydration agree.
  const storedTenant = useSyncExternalStore(
    () => () => {},
    () => localStorage.getItem(TENANT_KEY),
    () => null
  )
  const knownTenant = tenantId ?? storedTenant ?? undefined

  useEffect(() => {
    if (tenantId) localStorage.setItem(TENANT_KEY, tenantId)
  }, [tenantId])

  // A token in the URL exchanges immediately — the user just tapped their email link.
  useEffect(() => {
    if (token && exchange.isIdle) {
      exchange.mutate(token, { onSuccess: () => router.replace('/pay') })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token])

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!knownTenant || !email) return
    request.mutate({ tenantId: knownTenant, email })
  }

  return (
    <div className="mx-auto flex min-h-dvh w-full max-w-md flex-col justify-center px-4 py-10">
      <Card>
        <CardHeader>
          <div className="mb-2 inline-flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
            <ShieldCheck className="h-5 w-5 text-primary" aria-hidden />
          </div>
          <CardTitle>Manage your subscription</CardTitle>
          <CardDescription>
            {token && exchange.isPending
              ? 'Signing you in securely…'
              : 'Enter the email your subscription is under and we’ll send you a secure sign-in link.'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {token && exchange.isPending ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground" role="status">
              <Loader2 className="h-4 w-4 animate-spin" aria-hidden />
              Checking your link…
            </div>
          ) : request.isSuccess ? (
            <div className="flex items-start gap-3 rounded-lg border border-green-200 bg-green-50 p-4 text-sm text-green-800" role="status">
              <MailCheck className="mt-0.5 h-5 w-5 shrink-0" aria-hidden />
              <div>
                <p className="font-medium">Check your inbox</p>
                <p className="mt-1">
                  If an account exists for {email}, a secure link is on its way. It expires in 15
                  minutes and can be used once.
                </p>
              </div>
            </div>
          ) : (
            <form onSubmit={onSubmit} className="space-y-4">
              {exchange.isError && (
                <p className="rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800" role="alert">
                  That link is invalid or has expired. Request a fresh one below.
                </p>
              )}
              {request.isError && (
                <p className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800" role="alert">
                  We couldn&apos;t send the link — please try again shortly.
                </p>
              )}
              <div className="space-y-1.5">
                <label htmlFor="email" className="text-sm font-medium">
                  Email address
                </label>
                <Input
                  id="email"
                  type="email"
                  autoComplete="email"
                  required
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
              {!knownTenant && (
                <p className="text-sm text-muted-foreground">
                  This page needs the link your provider shared with you (it identifies whose
                  portal this is). Open the portal from your provider&apos;s site or email.
                </p>
              )}
              <Button type="submit" className="h-11 w-full" disabled={!knownTenant || request.isPending}>
                {request.isPending && <Loader2 className="h-4 w-4 animate-spin" aria-hidden />}
                Email me a secure link
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
      <p className="mt-4 text-center text-xs text-muted-foreground">
        Passwordless &amp; secure — we never ask you for a password.
      </p>
    </div>
  )
}
