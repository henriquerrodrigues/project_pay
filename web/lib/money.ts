const brl = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

// A API trabalha em centavos (int64).
export function formatBRL(cents: number): string {
  return brl.format(cents / 100);
}

// "12,50" ou "12.50" -> 1250. Retorna NaN se inválido.
export function parseBRLToCents(input: string): number {
  const value = Number(input.replace(",", "."));
  return Number.isFinite(value) ? Math.round(value * 100) : NaN;
}
