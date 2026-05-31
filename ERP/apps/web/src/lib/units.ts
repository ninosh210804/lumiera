/**
 * Unit conversion utilities for inventory management.
 * Supports: g, kg, ml, l, pcs (pieces/units)
 */

export type Unit = "g" | "kg" | "ml" | "l" | "pcs";

export const UNIT_LABELS: Record<Unit, string> = {
  g: "граммы (г)",
  kg: "килограммы (кг)",
  ml: "миллилитры (мл)",
  l: "литры (л)",
  pcs: "штуки (шт)",
};

export const UNIT_SHORT: Record<Unit, string> = {
  g: "г",
  kg: "кг",
  ml: "мл",
  l: "л",
  pcs: "шт",
};

// Conversion factors to base unit (g for weight, ml for volume, pcs for count)
const CONVERSION_FACTORS: Record<Unit, number> = {
  g: 1,
  kg: 1000,
  ml: 1,
  l: 1000,
  pcs: 1,
};

const UNIT_CATEGORIES: Record<Unit, "weight" | "volume" | "count"> = {
  g: "weight",
  kg: "weight",
  ml: "volume",
  l: "volume",
  pcs: "count",
};

/**
 * Convert quantity from one unit to another.
 * Only works within the same category (weight, volume, or count).
 * @param value - The quantity to convert
 * @param fromUnit - Source unit
 * @param toUnit - Target unit
 * @returns Converted quantity, or original value if units are incompatible
 */
export function convertUnit(value: number, fromUnit: Unit, toUnit: Unit): number {
  if (fromUnit === toUnit) return value;

  const fromCategory = UNIT_CATEGORIES[fromUnit];
  const toCategory = UNIT_CATEGORIES[toUnit];

  // Can only convert within same category
  if (fromCategory !== toCategory) {
    console.warn(`Cannot convert between ${fromUnit} and ${toUnit}: different categories`);
    return value; // Return original value
  }

  // Convert to base unit, then to target unit
  const baseValue = value * CONVERSION_FACTORS[fromUnit];
  const result = baseValue / CONVERSION_FACTORS[toUnit];

  // Round to 4 decimal places
  return Math.round(result * 10000) / 10000;
}

/**
 * Format a quantity with its unit for display.
 * @param qty - The quantity
 * @param unit - The unit
 * @param decimals - Number of decimal places to show
 * @returns Formatted string like "1.5 кг"
 */
export function formatQuantity(qty: number, unit: Unit, decimals: number = 2): string {
  const formatted = qty.toLocaleString("ru-RU", {
    minimumFractionDigits: 0,
    maximumFractionDigits: decimals,
  });
  return `${formatted} ${UNIT_SHORT[unit]}`;
}

/**
 * Get compatible units for conversion with a given unit.
 */
export function getCompatibleUnits(unit: Unit): Unit[] {
  const category = UNIT_CATEGORIES[unit];
  return (Object.keys(UNIT_CATEGORIES) as Unit[]).filter((u) => UNIT_CATEGORIES[u] === category);
}
