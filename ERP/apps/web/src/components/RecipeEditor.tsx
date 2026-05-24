import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";

interface Ingredient {
  id: string;
  name: string;
  unit: string;
  current_qty: number;
}

interface RecipeItem {
  id: string;
  ingredient_id?: string | null;
  ingredient_name?: string | null;
  qty: number;
  unit: string;
  unit_cost: number;
  line_cost: number;
}

interface RecipeDetail {
  id: string;
  name: string;
  items: RecipeItem[] | null;
  total_cost: number;
}

interface ProductLite {
  id: string;
  name: string;
  recipe_id: string | null;
}

/**
 * RecipeEditor manages the ingredient list (tech card) for a product.
 * On the first ingredient add it lazily creates a recipe and links it to the
 * product, so the user just picks ingredients and quantities. Selling the
 * product later deducts these quantities from warehouse stock.
 */
export default function RecipeEditor({
  product,
  locationId,
  categoryName,
  onClose,
}: {
  product: ProductLite;
  locationId: string;
  categoryName: string;
  onClose: () => void;
}) {
  const qc = useQueryClient();
  const [recipeId, setRecipeId] = useState<string | null>(product.recipe_id ?? null);
  const [pickId, setPickId] = useState("");
  const [qty, setQty] = useState("");
  const [error, setError] = useState("");

  const { data: ingredients = [] } = useQuery<Ingredient[]>({
    queryKey: ["ingredients", locationId],
    queryFn: () =>
      api.get("/ingredients", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: recipe, isLoading } = useQuery<RecipeDetail | null>({
    queryKey: ["recipe", recipeId],
    queryFn: () => api.get(`/recipes/${recipeId}`).then((r) => r.data.data),
    enabled: !!recipeId,
  });

  const items = recipe?.items ?? [];
  const usedIds = new Set(items.map((i) => i.ingredient_id));
  const available = ingredients.filter((i) => !usedIds.has(i.id));
  const selectedIng = ingredients.find((i) => i.id === pickId);

  function invalidate(rid: string) {
    qc.invalidateQueries({ queryKey: ["recipe", rid] });
    qc.invalidateQueries({ queryKey: ["products", locationId] });
  }

  // Ensures a recipe exists & is linked to the product; returns its id.
  async function ensureRecipe(): Promise<string> {
    if (recipeId) return recipeId;
    const created = await api.post("/recipes", {
      name: product.name,
      recipe_type: "product",
      yield_qty: 1,
      yield_unit: "pcs",
      notes: "",
    });
    const rid: string = created.data.data.id;
    await api.post(`/recipes/${rid}/link`, { product_id: product.id });
    setRecipeId(rid);
    return rid;
  }

  const addItem = useMutation({
    mutationFn: async () => {
      const rid = await ensureRecipe();
      await api.post(`/recipes/${rid}/items`, {
        ingredient_id: pickId,
        qty: Number(qty),
        unit: selectedIng?.unit ?? "g",
        sort_order: items.length,
      });
      return rid;
    },
    onSuccess: (rid) => {
      setPickId("");
      setQty("");
      invalidate(rid);
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      setError(msg ?? "Не удалось добавить ингредиент");
    },
  });

  const removeItem = useMutation({
    mutationFn: (itemId: string) => api.delete(`/recipes/${recipeId}/items/${itemId}`),
    onSuccess: () => recipeId && invalidate(recipeId),
  });

  const updateQty = useMutation({
    mutationFn: ({ item, newQty }: { item: RecipeItem; newQty: number }) =>
      api.put(`/recipes/${recipeId}/items/${item.id}`, {
        qty: newQty,
        unit: item.unit,
        sort_order: 0,
      }),
    onSuccess: () => recipeId && invalidate(recipeId),
  });

  function submitAdd(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!pickId) return setError("Выберите ингредиент");
    if (!qty || Number(qty) <= 0 || Number.isNaN(Number(qty))) return setError("Введите количество");
    addItem.mutate();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div
        className="bg-white rounded-2xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <div>
            <h2 className="text-lg font-bold text-gray-900">Состав: {product.name}</h2>
            <p className="text-xs text-gray-400">{categoryName} · списывается со склада при продаже</p>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl leading-none">×</button>
        </div>

        <div className="p-6 space-y-4">
          {/* Current ingredients */}
          {isLoading ? (
            <div className="h-20 bg-gray-100 rounded-xl animate-pulse" />
          ) : items.length === 0 ? (
            <div className="text-center text-sm text-gray-400 py-6">
              Ингредиенты ещё не добавлены.
              <br />
              Добавьте, что расходуется на 1 порцию.
            </div>
          ) : (
            <div className="space-y-2">
              {items.map((it) => (
                <div key={it.id} className="flex items-center gap-2 bg-gray-50 rounded-xl px-3 py-2">
                  <span className="flex-1 text-sm font-medium text-gray-800">{it.ingredient_name}</span>
                  <input
                    type="number"
                    min="0"
                    step="any"
                    defaultValue={it.qty}
                    onBlur={(e) => {
                      const v = Number(e.target.value);
                      if (v > 0 && v !== it.qty) updateQty.mutate({ item: it, newQty: v });
                    }}
                    className="w-20 px-2 py-1 text-sm text-right border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand/30"
                  />
                  <span className="w-8 text-xs text-gray-400">{it.unit}</span>
                  <button
                    onClick={() => removeItem.mutate(it.id)}
                    className="text-red-400 hover:text-red-600 text-lg leading-none px-1"
                    title="Убрать"
                  >
                    ×
                  </button>
                </div>
              ))}
              <div className="flex justify-between pt-1 text-xs text-gray-500">
                <span>Себестоимость порции</span>
                <span className="font-semibold tabular-nums">
                  {(recipe?.total_cost ?? 0).toLocaleString("ru-RU", { maximumFractionDigits: 2 })} ₸
                </span>
              </div>
            </div>
          )}

          {/* Add ingredient */}
          <form onSubmit={submitAdd} className="border-t border-gray-100 pt-4 space-y-3">
            <div className="flex gap-2">
              <select
                value={pickId}
                onChange={(e) => setPickId(e.target.value)}
                className="flex-1 px-3 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
              >
                <option value="">Выберите ингредиент…</option>
                {available.map((i) => (
                  <option key={i.id} value={i.id}>
                    {i.name} (на складе: {i.current_qty} {i.unit})
                  </option>
                ))}
              </select>
              <input
                type="number"
                min="0"
                step="any"
                value={qty}
                onChange={(e) => setQty(e.target.value)}
                placeholder="Кол-во"
                className="w-24 px-3 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
              />
              <span className="self-center w-8 text-xs text-gray-400">{selectedIng?.unit ?? ""}</span>
            </div>

            {error && <p className="text-red-500 text-sm bg-red-50 px-3 py-2 rounded-lg">{error}</p>}
            {ingredients.length === 0 && (
              <p className="text-amber-600 text-xs">
                Нет ингредиентов. Сначала добавьте их на странице «Склад».
              </p>
            )}

            <button
              type="submit"
              disabled={addItem.isPending}
              className="w-full bg-brand text-white py-2.5 rounded-lg text-sm font-semibold hover:bg-brand-600 transition disabled:opacity-50"
            >
              {addItem.isPending ? "Добавление…" : "+ Добавить ингредиент"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
