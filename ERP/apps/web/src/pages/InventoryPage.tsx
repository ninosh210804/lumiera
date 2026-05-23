import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { format } from "date-fns";
import { ru } from "date-fns/locale";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

interface IngredientDTO {
  id: string;
  name: string;
  unit: string;
  is_perishable: boolean;
  min_stock_alert: number;
  current_avg_cost: number;
  current_qty: number;
  is_active: boolean;
}

interface StockMovementDTO {
  id: string;
  ingredient_id: string;
  ingredient_name?: string;
  reason: string;
  qty_delta: number;
  unit_cost: number;
  note?: string;
  created_at: string;
}

const UNIT_LABELS: Record<string, string> = {
  g: "г", kg: "кг", ml: "мл", l: "л", pcs: "шт",
};

const MOVEMENT_LABELS: Record<string, { label: string; cls: string }> = {
  purchase:   { label: "Приход",       cls: "text-emerald-600" },
  sale:       { label: "Продажа",      cls: "text-blue-600" },
  waste:      { label: "Отход",        cls: "text-red-600" },
  spill:      { label: "Пролив",       cls: "text-orange-600" },
  count:      { label: "Инвентаризация", cls: "text-purple-600" },
};

type Tab = "stock" | "movements";

export default function InventoryPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const [tab, setTab] = useState<Tab>("stock");
  const [search, setSearch] = useState("");

  const { data: ingredients = [], isLoading } = useQuery<IngredientDTO[]>({
    queryKey: ["ingredients", locationId],
    queryFn: () =>
      api.get("/ingredients", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: movements = [], isLoading: loadingMov } = useQuery<StockMovementDTO[]>({
    queryKey: ["movements", locationId],
    queryFn: () =>
      api.get("/stock/movements", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: !!locationId && tab === "movements",
  });

  const lowStock = ingredients.filter((i) => i.current_qty <= i.min_stock_alert && i.min_stock_alert > 0);
  const filtered = ingredients.filter((i) =>
    i.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Склад</h1>
          {lowStock.length > 0 && (
            <p className="text-sm text-amber-600 mt-0.5 font-medium">
              ⚠️ {lowStock.length} позиц. ниже минимума
            </p>
          )}
        </div>

        <div className="flex gap-1 bg-gray-100 p-1 rounded-xl">
          {(["stock", "movements"] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={[
                "px-4 py-1.5 rounded-lg text-sm font-medium transition-all",
                tab === t ? "bg-white text-brand shadow-sm" : "text-gray-500 hover:text-gray-700",
              ].join(" ")}
            >
              {t === "stock" ? "Остатки" : "Движение"}
            </button>
          ))}
        </div>
      </div>

      {tab === "stock" && (
        <>
          <div className="mb-4">
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Поиск товара..."
              className="w-full px-4 py-2 border border-gray-200 rounded-lg"
            />
          </div>

          {lowStock.length > 0 && (
            <div className="mb-6 bg-amber-50 border border-amber-200 rounded-lg p-4">
              <h3 className="font-semibold text-amber-900 mb-3">
                ⚠️ Ниже минимума ({lowStock.length} позиц.)
              </h3>
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                {lowStock.map((i) => (
                  <div key={i.id} className="bg-white rounded p-2 text-center">
                    <div className="font-semibold text-sm text-gray-900 truncate">{i.name}</div>
                    <div className="text-red-600 text-xs font-bold">
                      {i.current_qty.toFixed(2)} {UNIT_LABELS[i.unit]}
                    </div>
                    <div className="text-gray-500 text-[10px]">мин: {i.min_stock_alert.toFixed(2)}</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            {isLoading ? (
              <div className="p-8 text-center text-gray-500">Загрузка...</div>
            ) : filtered.length === 0 ? (
              <div className="p-8 text-center text-gray-500">Товары не найдены</div>
            ) : (
              <table className="w-full text-sm">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left font-semibold text-gray-900">Товар</th>
                    <th className="px-4 py-3 text-right font-semibold text-gray-900">Кол-во</th>
                    <th className="px-4 py-3 text-right font-semibold text-gray-900">Мин. уровень</th>
                    <th className="px-4 py-3 text-right font-semibold text-gray-900">Ср. стоимость</th>
                    <th className="px-4 py-3 text-left font-semibold text-gray-900">Статус</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {filtered.map((i) => {
                    const isLow = i.current_qty <= i.min_stock_alert && i.min_stock_alert > 0;
                    return (
                      <tr key={i.id} className={isLow ? "bg-amber-50" : "hover:bg-gray-50"}>
                        <td className="px-4 py-3 font-medium text-gray-900">{i.name}</td>
                        <td className="px-4 py-3 text-right">
                          <span className={isLow ? "text-red-600 font-bold" : ""}>
                            {i.current_qty.toFixed(2)} {UNIT_LABELS[i.unit]}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right text-gray-600">{i.min_stock_alert.toFixed(2)}</td>
                        <td className="px-4 py-3 text-right text-gray-600">
                          {i.current_avg_cost.toLocaleString("ru-RU")} ₸
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-block px-2 py-1 rounded text-xs font-medium ${
                              isLow
                                ? "bg-red-100 text-red-700"
                                : i.current_qty === 0
                                ? "bg-gray-100 text-gray-700"
                                : "bg-green-100 text-green-700"
                            }`}
                          >
                            {isLow ? "⚠️ Низко" : i.current_qty === 0 ? "❌ Нет" : "✅ OK"}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            )}
          </div>
        </>
      )}

      {tab === "movements" && (
        <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
          {loadingMov ? (
            <div className="p-8 text-center text-gray-500">Загрузка...</div>
          ) : movements.length === 0 ? (
            <div className="p-8 text-center text-gray-500">История не найдена</div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left font-semibold text-gray-900">Товар</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-900">Операция</th>
                  <th className="px-4 py-3 text-right font-semibold text-gray-900">Кол-во</th>
                  <th className="px-4 py-3 text-right font-semibold text-gray-900">Стоимость ед.</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-900">Дата/время</th>
                  <th className="px-4 py-3 text-left font-semibold text-gray-900">Примечание</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {movements.map((m) => {
                  const isIncoming = m.qty_delta > 0;
                  const mvLabel = MOVEMENT_LABELS[m.reason] || { label: m.reason, cls: "text-gray-600" };
                  return (
                    <tr key={m.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 font-medium text-gray-900">{m.ingredient_name}</td>
                      <td className="px-4 py-3">
                        <span className={`inline-block px-2 py-1 rounded text-xs font-medium ${mvLabel.cls}`}>
                          {mvLabel.label}
                        </span>
                      </td>
                      <td
                        className={`px-4 py-3 text-right font-medium ${
                          isIncoming ? "text-green-600" : "text-red-600"
                        }`}
                      >
                        {isIncoming ? "+" : "−"}{Math.abs(m.qty_delta).toFixed(2)}
                      </td>
                      <td className="px-4 py-3 text-right text-gray-600">
                        {m.unit_cost.toLocaleString("ru-RU")} ₸
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600">
                        {format(new Date(m.created_at), "dd MMM yyyy, HH:mm", { locale: ru })}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600 max-w-xs truncate">
                        {m.note || "—"}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  );
}
