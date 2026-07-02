'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { api } from '@/lib/api'
import type { components } from '@/lib/api/v1'

export type Subscription = components['schemas']['Subscription']
export type Invoice = components['schemas']['Invoice']
export type InvoiceDetail = components['schemas']['InvoiceDetail']
export type PortalContext = components['schemas']['PortalContext']

// ApiError carries the backend's error envelope; request_id is surfaced in error
// states so a failed action is traceable to server logs.
export class ApiError extends Error {
  status: number
  requestId?: string
  constructor(status: number, message: string, requestId?: string) {
    super(message)
    this.status = status
    this.requestId = requestId
  }
}

function toApiError(status: number, err: unknown): ApiError {
  const e = err as { message?: string; request_id?: string } | undefined
  return new ApiError(status, e?.message || 'something went wrong', e?.request_id)
}

// The contract only declares success statuses for some endpoints, which makes
// openapi-fetch type `error` as never and lets TS narrow `response` away entirely.
// Widen to a plain Response and gate on `ok` instead.
function ensureOk(response: Response, error: unknown) {
  if (!response.ok) throw toApiError(response.status, error)
}

export function usePortalMe() {
  return useQuery({
    queryKey: ['portal', 'me'],
    retry: false,
    queryFn: async () => {
      const { data, error, response } = await api.GET('/portal/me')
      ensureOk(response as Response, error)
      return data
    },
  })
}

// null means "signed in, but no subscription yet" — a first-class empty state.
export function useSubscription() {
  return useQuery({
    queryKey: ['portal', 'subscription'],
    queryFn: async (): Promise<Subscription | null> => {
      const { data, error, response } = await api.GET('/portal/subscription')
      const res = response as Response
      if (res.status === 404) return null
      ensureOk(res, error)
      return data ?? null
    },
  })
}

export function useInvoices() {
  return useQuery({
    queryKey: ['portal', 'invoices'],
    queryFn: async () => {
      const { data, error, response } = await api.GET('/portal/invoices')
      ensureOk(response as Response, error)
      return data ?? []
    },
  })
}

export function useInvoice(id: string) {
  return useQuery({
    queryKey: ['portal', 'invoices', id],
    queryFn: async () => {
      const { data, error, response } = await api.GET('/portal/invoices/{id}', {
        params: { path: { id } },
      })
      ensureOk(response as Response, error)
      return data
    },
  })
}

export function usePaymentMethod() {
  return useQuery({
    queryKey: ['portal', 'payment-method'],
    queryFn: async () => {
      const { data, error, response } = await api.GET('/portal/payment-method')
      ensureOk(response as Response, error)
      return data
    },
  })
}

export function useRequestAccess() {
  return useMutation({
    mutationFn: async (vars: { tenantId: string; email: string }) => {
      const { error, response } = await api.POST('/portal/access-request', {
        body: { tenant_id: vars.tenantId, email: vars.email },
      })
      ensureOk(response as Response, error)
    },
  })
}

export function useExchangeToken() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (token: string) => {
      const { data, error, response } = await api.POST('/portal/session', {
        body: { token },
      })
      ensureOk(response as Response, error)
      return data
    },
    onSuccess: (ctx) => qc.setQueryData(['portal', 'me'], ctx),
  })
}

export function useCancelSubscription() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (atPeriodEnd: boolean) => {
      const { data, error, response } = await api.POST('/portal/subscription/cancel', {
        body: { at_period_end: atPeriodEnd },
      })
      ensureOk(response as Response, error)
      return data
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portal', 'subscription'] }),
  })
}

export function usePortalLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      await api.POST('/portal/logout')
    },
    onSuccess: () => qc.clear(),
  })
}
