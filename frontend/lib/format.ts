import { format } from 'date-fns'

export function naira(amountInMinor: number | bigint): string {
  const amountInMajor = Number(amountInMinor) / 100
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: 'NGN',
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amountInMajor)
}

// "30 Jun 2026" per the design brief; renewal copy is "Renews {formatDate(...)}".
export function formatDate(iso: string | Date): string {
  return format(new Date(iso), 'd MMM yyyy')
}

export function intervalLabel(interval: string | undefined): string {
  return interval === 'year' ? '/yr' : '/mo'
}
