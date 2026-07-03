export function naira(minorUnits: number): string {
  // Minor units are kobo. So we divide by 100 to get major units (Naira).
  const majorUnits = minorUnits / 100;
  
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: 'NGN',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(majorUnits);
}
