const KZT_FORMATTER = new Intl.NumberFormat("ru-KZ", {
  style: "currency",
  currency: "KZT",
  maximumFractionDigits: 0,
});

export function formatKZT(amount: number): string {
  return KZT_FORMATTER.format(amount);
}

export function roundKZT(amount: number): number {
  return Math.round(amount);
}

export function calcMarginPct(revenue: number, cost: number): number {
  if (revenue === 0) return 0;
  return Math.round(((revenue - cost) / revenue) * 100 * 10) / 10;
}
