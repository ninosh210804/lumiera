import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { format } from "date-fns";
import { ru } from "date-fns/locale";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

interface IngredientDTO {
  id: string;
  name: string;
  unit: string;
  current_qty: number;
  min_stock_alert: number;
  current_avg_cost: number;
  is_perishable: boolean;
  default_shelf_life_days?: number | null;
  is_active: boolean;
}

interface SupplierDTO {
  id: string;
  name: string;
  contact: string;
  phone: string;
}

interface StockMovementDTO {
  id: string;
  ingredient_id: string;
  ingredient_name: string;
  qty_delta: number;
  unit_cost: number;
  reason: string;
  note?: string;
  created_at: string;
}

interface RefillItem {
  ingredient_id: string;
  qty: number;
  unit_cost: number;
  expires_at?: string;
}

export default function WarehousePage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const queryClient = useQueryClient();

  const [tab, setTab] = useState<"refill" | "history">("refill");
  const [selectedSupplierId, setSelectedSupplierId] = useState<string>("");
  const [refillItems, setRefillItems] = useState<RefillItem[]>([]);
  const [notes, setNotes] = useState("");
  const [showNewIngredient, setShowNewIngredient] = useState(false);
  const [editIngredient, setEditIngredient] = useState<IngredientDTO | null>(null);

  const { data: ingredients = [] } = useQuery<IngredientDTO[]>({
    queryKey: ["ingredients-warehouse", locationId],
    queryFn: () =>
      api.get("/ingredients", { params: { location_id: locationId } })
        .then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: suppliers = [] } = useQuery<SupplierDTO[]>({
    queryKey: ["suppliers"],
    queryFn: () =>
      api.get("/suppliers")
        .then((r) => r.data.data ?? []),
  });

  const { data: movements = [] } = useQuery<StockMovementDTO[]>({
    queryKey: ["stock-movements-warehouse", locationId, tab],
    queryFn: () =>
      api.get("/stock/movements", { params: { location_id: locationId } })
        .then((r) => r.data.data ?? []),
    enabled: !!locationId && tab === "history",
  });

  const receiveMutation = useMutation({
    mutationFn: (data: {
      location_id: string;
      supplier_id?: string;
      notes: string;
      items: RefillItem[];
    }) => api.post("/stock/receive", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ingredients-warehouse"] });
      queryClient.invalidateQueries({ queryKey: ["stock-movements-warehouse"] });
      setRefillItems([]);
      setSelectedSupplierId("");
      setNotes("");
      alert("Товар успешно принят!");
    },
    onError: (err) => {
      alert(`Ошибка: ${err instanceof Error ? err.message : "Не удалось принять товар"}`);
    },
  });

  const addRefillItem = (ingredientId: string) => {
    const ingredient = ingredients.find((i) => i.id === ingredientId);
    if (!ingredient) return;

    if (refillItems.some((item) => item.ingredient_id === ingredientId)) {
      alert("Этот товар уже добавлен");
      return;
    }

    setRefillItems([
      ...refillItems,
      {
        ingredient_id: ingredientId,
        qty: 0,
        unit_cost: ingredient.current_avg_cost || 0,
        expires_at: undefined,
      },
    ]);
  };

  const updateRefillItem = (idx: number, field: string, value: any) => {
    const updated = [...refillItems];
    updated[idx] = { ...updated[idx], [field]: value };
    setRefillItems(updated);
  };

  const removeRefillItem = (idx: number) => {
    setRefillItems(refillItems.filter((_, i) => i !== idx));
  };

  const handleReceiveStock = () => {
    if (refillItems.length === 0) {
      alert("Добавьте товар для приема");
      return;
    }

    if (refillItems.some((item) => item.qty <= 0)) {
      alert("Количество должно быть больше 0");
      return;
    }

    receiveMutation.mutate({
      location_id: locationId,
      supplier_id: selectedSupplierId || undefined,
      notes,
      items: refillItems,
    });
  };

  const lowStockIngredients = ingredients.filter(
    (i) => i.current_qty <= i.min_stock_alert && i.min_stock_alert > 0
  );

  const totalRefillCost = refillItems.reduce(
    (sum, item) => sum + item.qty * item.unit_cost,
    0
  );

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Управление складом</h1>

      {/* Tabs */}
      <div className="flex gap-2 mb-6 border-b border-gray-200">
        <button
          onClick={() => setTab("refill")}
          className={`px-4 py-3 font-medium border-b-2 transition ${
            tab === "refill"
              ? "border-blue-600 text-blue-600"
              : "border-transparent text-gray-500 hover:text-gray-700"
          }`}
        >
          Приход товара
        </button>
        <button
          onClick={() => setTab("history")}
          className={`px-4 py-3 font-medium border-b-2 transition ${
            tab === "history"
              ? "border-blue-600 text-blue-600"
              : "border-transparent text-gray-500 hover:text-gray-700"
          }`}
        >
          История движения
        </button>
      </div>

      {tab === "refill" ? (
        <div className="grid grid-cols-3 gap-6">
          {/* Left: Low Stock Alerts & Ingredient List */}
          <div className="col-span-1">
            {lowStockIngredients.length > 0 && (
              <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-lg">
                <h3 className="font-semibold text-amber-900 mb-2">
                  ⚠️ Ниже минимума ({lowStockIngredients.length})
                </h3>
                <div className="space-y-1">
                  {lowStockIngredients.map((ing) => (
                    <button
                      key={ing.id}
                      onClick={() => addRefillItem(ing.id)}
                      className="w-full text-left text-sm px-2 py-1 hover:bg-amber-100 rounded text-amber-900"
                    >
                      {ing.name}: {ing.current_qty.toFixed(2)} {ing.unit}
                    </button>
                  ))}
                </div>
              </div>
            )}

            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="font-semibold text-gray-900">Все товары</h3>
                <button
                  onClick={() => setShowNewIngredient(true)}
                  className="text-xs font-semibold px-2.5 py-1.5 rounded-lg bg-blue-600 text-white hover:bg-blue-500 transition"
                >
                  + Новый товар
                </button>
              </div>
              <div className="space-y-1 max-h-96 overflow-y-auto">
                {ingredients.map((ing) => (
                  <div
                    key={ing.id}
                    className="group flex items-center gap-1 rounded hover:bg-gray-100 transition"
                  >
                    <button
                      onClick={() => addRefillItem(ing.id)}
                      disabled={refillItems.some((item) => item.ingredient_id === ing.id)}
                      className="flex-1 text-left text-sm px-2 py-1.5 disabled:opacity-50 rounded transition"
                    >
                      <div className="font-medium text-gray-900">{ing.name}</div>
                      <div className="text-xs text-gray-500">
                        Осталось: {ing.current_qty.toFixed(2)} {ing.unit}
                      </div>
                    </button>
                    <button
                      onClick={() => setEditIngredient(ing)}
                      title="Редактировать"
                      className="opacity-0 group-hover:opacity-100 shrink-0 px-2 py-1 text-gray-400 hover:text-blue-600 transition"
                    >
                      ✏️
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Right: Refill Form */}
          <div className="col-span-2">
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              {/* Supplier Selection */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Поставщик (опционально)
                </label>
                <select
                  value={selectedSupplierId}
                  onChange={(e) => setSelectedSupplierId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                >
                  <option value="">Не выбран</option>
                  {suppliers.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Notes */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Примечания
                </label>
                <textarea
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  placeholder="Номер счета, дата доставки, и т.д."
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                  rows={2}
                />
              </div>

              {/* Refill Items List */}
              <div className="mb-6">
                <h3 className="font-semibold text-gray-900 mb-3">Товары для приема</h3>
                {refillItems.length === 0 ? (
                  <div className="text-center text-gray-500 py-6">
                    Выберите товары из списка слева
                  </div>
                ) : (
                  <div className="space-y-3">
                    {refillItems.map((item, idx) => {
                      const ingredient = ingredients.find((i) => i.id === item.ingredient_id);
                      return (
                        <div
                          key={idx}
                          className="border border-gray-200 rounded-lg p-3 bg-gray-50"
                        >
                          <div className="flex justify-between items-start mb-2">
                            <div>
                              <div className="font-medium text-gray-900">
                                {ingredient?.name}
                              </div>
                              <div className="text-xs text-gray-500">
                                Ед. изм.: {ingredient?.unit}
                              </div>
                            </div>
                            <button
                              onClick={() => removeRefillItem(idx)}
                              className="text-red-600 hover:text-red-700 text-sm font-medium"
                            >
                              Удалить
                            </button>
                          </div>

                          <div className="grid grid-cols-3 gap-2">
                            <div>
                              <label className="text-xs text-gray-600 mb-1 block">
                                Кол-во
                              </label>
                              <input
                                type="number"
                                value={item.qty}
                                onChange={(e) =>
                                  updateRefillItem(idx, "qty", parseFloat(e.target.value) || 0)
                                }
                                step="0.1"
                                min="0"
                                className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                              />
                            </div>
                            <div>
                              <label className="text-xs text-gray-600 mb-1 block">
                                Цена/шт (₸)
                              </label>
                              <input
                                type="number"
                                value={item.unit_cost}
                                onChange={(e) =>
                                  updateRefillItem(idx, "unit_cost", parseFloat(e.target.value) || 0)
                                }
                                step="0.01"
                                min="0"
                                className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                              />
                            </div>
                            <div>
                              <label className="text-xs text-gray-600 mb-1 block">
                                Сумма
                              </label>
                              <div className="px-2 py-1 bg-white border border-gray-300 rounded text-sm font-medium">
                                {(item.qty * item.unit_cost).toLocaleString("ru-RU")}
                              </div>
                            </div>
                          </div>

                          {ingredient?.is_perishable && (
                            <div className="mt-2">
                              <label className="text-xs text-gray-600 mb-1 block">
                                Годен до (опционально)
                              </label>
                              <input
                                type="date"
                                value={item.expires_at || ""}
                                onChange={(e) =>
                                  updateRefillItem(idx, "expires_at", e.target.value || undefined)
                                }
                                className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                              />
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>

              {/* Totals & Submit */}
              <div className="border-t border-gray-200 pt-4">
                <div className="flex justify-between items-center mb-4">
                  <span className="text-lg font-semibold text-gray-900">
                    Общая сумма:
                  </span>
                  <span className="text-2xl font-bold text-blue-600">
                    {totalRefillCost.toLocaleString("ru-RU")} ₸
                  </span>
                </div>
                <button
                  onClick={handleReceiveStock}
                  disabled={receiveMutation.isPending || refillItems.length === 0}
                  className="w-full bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-400 text-white font-bold py-3 rounded-lg transition"
                >
                  {receiveMutation.isPending ? "Обработка..." : "Принять товар"}
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : (
        /* History Tab */
        <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">
                  Товар
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">
                  Операция
                </th>
                <th className="px-6 py-3 text-right text-sm font-semibold text-gray-900">
                  Количество
                </th>
                <th className="px-6 py-3 text-right text-sm font-semibold text-gray-900">
                  Стоимость
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">
                  Дата
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">
                  Примечание
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {movements.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                    История не найдена
                  </td>
                </tr>
              ) : (
                movements.map((m) => (
                  <tr key={m.id} className="hover:bg-gray-50 transition">
                    <td className="px-6 py-3 text-sm font-medium text-gray-900">
                      {m.ingredient_name}
                    </td>
                    <td className="px-6 py-3 text-sm">
                      <span
                        className={`inline-block px-2 py-1 rounded text-xs font-medium ${
                          m.reason === "purchase"
                            ? "bg-emerald-50 text-emerald-700"
                            : m.reason === "sale"
                            ? "bg-blue-50 text-blue-700"
                            : m.reason === "waste"
                            ? "bg-red-50 text-red-700"
                            : "bg-gray-50 text-gray-700"
                        }`}
                      >
                        {m.reason === "purchase"
                          ? "Приход"
                          : m.reason === "sale"
                          ? "Продажа"
                          : m.reason === "waste"
                          ? "Отход"
                          : m.reason}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-sm text-right font-medium">
                      {m.qty_delta > 0 ? "+" : ""}{m.qty_delta.toFixed(2)}
                    </td>
                    <td className="px-6 py-3 text-sm text-right text-gray-600">
                      {(Math.abs(m.qty_delta) * (m.unit_cost ?? 0)).toLocaleString("ru-RU")} ₸
                    </td>
                    <td className="px-6 py-3 text-sm text-gray-600">
                      {format(new Date(m.created_at), "dd MMM yyyy, HH:mm", {
                        locale: ru,
                      })}
                    </td>
                    <td className="px-6 py-3 text-sm text-gray-600 max-w-xs truncate">
                      {m.note || "—"}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {showNewIngredient && (
        <IngredientModal
          onClose={() => setShowNewIngredient(false)}
          onSaved={() => {
            queryClient.invalidateQueries({ queryKey: ["ingredients-warehouse"] });
            setShowNewIngredient(false);
          }}
        />
      )}

      {editIngredient && (
        <IngredientModal
          ingredient={editIngredient}
          onClose={() => setEditIngredient(null)}
          onSaved={() => {
            queryClient.invalidateQueries({ queryKey: ["ingredients-warehouse"] });
            setEditIngredient(null);
          }}
        />
      )}
    </div>
  );
}

const UNIT_OPTIONS: { value: string; label: string }[] = [
  { value: "g", label: "граммы (g)" },
  { value: "kg", label: "килограммы (kg)" },
  { value: "ml", label: "миллилитры (ml)" },
  { value: "l", label: "литры (l)" },
  { value: "pcs", label: "штуки (pcs)" },
];

function IngredientModal({
  ingredient,
  onClose,
  onSaved,
}: {
  ingredient?: IngredientDTO;
  onClose: () => void;
  onSaved: () => void;
}) {
  const isEdit = !!ingredient;
  const [name, setName] = useState(ingredient?.name ?? "");
  const [unit, setUnit] = useState(ingredient?.unit ?? "g");
  const [minStock, setMinStock] = useState(
    ingredient?.min_stock_alert != null ? String(ingredient.min_stock_alert) : ""
  );
  const [perishable, setPerishable] = useState(ingredient?.is_perishable ?? false);
  const [shelfLife, setShelfLife] = useState(
    ingredient?.default_shelf_life_days != null ? String(ingredient.default_shelf_life_days) : ""
  );
  const [error, setError] = useState("");

  const save = useMutation({
    mutationFn: () => {
      const body: Record<string, unknown> = {
        name: name.trim(),
        unit,
        is_perishable: perishable,
        min_stock_alert: minStock ? Number(minStock) : 0,
        default_shelf_life_days: perishable && shelfLife ? Number(shelfLife) : null,
      };
      if (isEdit) {
        body.is_active = ingredient!.is_active;
        return api.put(`/ingredients/${ingredient!.id}`, body);
      }
      return api.post("/ingredients", body);
    },
    onSuccess: onSaved,
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      setError(msg ?? (isEdit ? "Не удалось сохранить товар" : "Не удалось создать товар"));
    },
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!name.trim()) return setError("Введите название");
    save.mutate();
  }

  const field = "w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500 transition";
  const label = "block text-sm font-medium text-gray-700 mb-1";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div className="bg-white rounded-2xl shadow-xl w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <h2 className="text-lg font-bold text-gray-900">
            {isEdit ? "Редактировать товар" : "Новый товар на склад"}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl leading-none">×</button>
        </div>
        <form onSubmit={submit} className="p-6 space-y-4">
          <div>
            <label className={label}>Название</label>
            <input className={field} value={name} onChange={(e) => setName(e.target.value)} placeholder="Молоко цельное" autoFocus />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className={label}>Единица</label>
              <select className={field} value={unit} onChange={(e) => setUnit(e.target.value)}>
                {UNIT_OPTIONS.map((u) => (
                  <option key={u.value} value={u.value}>{u.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className={label}>Мин. остаток</label>
              <input className={field} type="number" min="0" step="any" value={minStock} onChange={(e) => setMinStock(e.target.value)} placeholder="0" />
            </div>
          </div>
          <label className="flex items-center gap-2 text-sm text-gray-700">
            <input type="checkbox" checked={perishable} onChange={(e) => setPerishable(e.target.checked)} className="rounded" />
            Скоропортящийся
          </label>
          {perishable && (
            <div>
              <label className={label}>Срок годности (дней)</label>
              <input className={field} type="number" min="0" value={shelfLife} onChange={(e) => setShelfLife(e.target.value)} placeholder="7" />
            </div>
          )}
          {error && <p className="text-red-500 text-sm bg-red-50 px-3 py-2 rounded-lg">{error}</p>}
          <button
            type="submit"
            disabled={save.isPending}
            className="w-full bg-blue-600 text-white py-2.5 rounded-lg text-sm font-semibold hover:bg-blue-500 transition disabled:opacity-50"
          >
            {save.isPending
              ? (isEdit ? "Сохранение…" : "Создание…")
              : (isEdit ? "Сохранить" : "Создать товар")}
          </button>
        </form>
      </div>
    </div>
  );
}
