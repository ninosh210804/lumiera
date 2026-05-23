export interface Category {
  id: string;
  locationId: string;
  name: string;
  icon: string;
  sortOrder: number;
  isActive: boolean;
}

export interface ModifierOption {
  id: string;
  groupId: string;
  name: string;
  priceDelta: number;
  linkedIngredientId?: string;
  ingredientQtyDelta: number;
  isActive: boolean;
  sortOrder: number;
}

export interface ModifierGroup {
  id: string;
  productId: string;
  name: string;
  selectionType: "single" | "multi";
  required: boolean;
  minSelect: number;
  maxSelect: number;
  sortOrder: number;
  options: ModifierOption[];
}

export interface Product {
  id: string;
  locationId: string;
  categoryId: string;
  recipeId?: string;
  name: string;
  description: string;
  sku?: string;
  basePrice: number;
  isActive: boolean;
  isStopListed: boolean;
  imageUrl?: string;
  sortOrder: number;
  categoryName: string;
  modifierGroups?: ModifierGroup[];
}
