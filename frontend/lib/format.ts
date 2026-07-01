export function naira(amountInMinor: number | bigint): string {
  const amountInMajor = Number(amountInMinor) / 100
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: 'NGN',
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amountInMajor)
}
