export type StockReason = "sale" | "waste" | "spill" | "gift" | "staff" | "count" | "purchase" | "adjustment";
export type CostMethod = "AVG" | "FIFO";

export interface Ingredient {
  id: string;
  locationId: string;
  name: string;
  unit: "g" | "kg" | "ml" | "l" | "pcs";
  isPerishable: boolean;
  defaultShelfLifeDays?: number;
  minStockAlert: number;
  currentAvgCost: number;
  currentQty: number;
  isActive: boolean;
}

export interface RecipeItem {
  id: string;
  recipeId: string;
  ingredientId?: string;
  subRecipeId?: string;
  qty: number;
  unit: string;
  ingredientName?: string;
  currentAvgCost?: number;
  currentQty?: number;
}

export interface Recipe {
  id: string;
  locationId: string;
  name: string;
  recipeType: "product" | "semi_finished";
  yieldQty: number;
  yieldUnit: string;
  items: RecipeItem[];
}

export interface StockMovement {
  id: string;
  locationId: string;
  ingredientId: string;
  qtyDelta: number;
  unitCostSnapshot: number;
  reason: StockReason;
  orderId?: string;
  note?: string;
  clientUuid: string;
  createdAt: string;
}
