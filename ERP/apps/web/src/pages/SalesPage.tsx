import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

interface SaleEvent {
  id: string;
  name: string;
  is_active: boolean;
  item_count: number;
}

interface SaleItem {
  product_id: string;
  product_name: string;
  base_price: number;
  sale_price: number;
}

interface SaleDetail extends SaleEvent {
  items: SaleItem[] | null;
}

interface ProductDTO {
  id: string;
  name: string;
  base_price: number;
  category_name: string;
}

export default function SalesPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const qc = useQueryClient();
  const [newName, setNewName] = useState("");
  const [openId, setOpenId] = useState<string | null>(null);

  const { data: events = [] } = useQuery<SaleEvent[]>({
    queryKey: ["sales", locationId],
    queryFn: () => api.get("/sales", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const create = useMutation({
    mutationFn: () => api.post("/sales", { name: newName.trim(), is_active: true }),
    onSuccess: () => {
      setNewName("");
      qc.invalidateQueries({ queryKey: ["sales", locationId] });
    },
  });

  const toggle = useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) =>
      api.post(`/sales/${id}/active`, { is_active: active }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["sales", locationId] });
      qc.invalidateQueries({ queryKey: ["products"] });
    },
  });

  const del = useMutation({
    mutationFn: (id: string) => api.delete(`/sales/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sales", locationId] }),
  });

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Акции / распродажи</h1>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (newName.trim()) create.mutate();
        }}
        className="flex gap-2 mb-6"
      >
        <input
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          placeholder="Название акции (напр. «Счастливые часы»)"
          className="flex-1 px-3 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
        />
        <button
          type="submit"
          disabled={create.isPending}
          className="px-4 py-2.5 rounded-lg bg-brand text-white text-sm font-semibold hover:bg-brand-600 transition disabled:opacity-50"
        >
          + Создать акцию
        </button>
      </form>

      {events.length === 0 ? (
        <div className="text-center text-gray-300 py-16">
          <div className="text-4xl mb-2">🏷️</div>
          <p className="text-sm">Акций пока нет</p>
        </div>
      ) : (
        <div className="space-y-3">
          {events.map((ev) => (
            <div key={ev.id} className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
              <div className="flex items-center justify-between px-5 py-3.5">
                <button
                  onClick={() => setOpenId(openId === ev.id ? null : ev.id)}
                  className="flex items-center gap-3 text-left"
                >
                  <span className="text-gray-400">{openId === ev.id ? "▾" : "▸"}</span>
                  <div>
                    <p className="font-semibold text-gray-900">{ev.name}</p>
                    <p className="text-xs text-gray-400">{ev.item_count} товаров</p>
                  </div>
                </button>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => toggle.mutate({ id: ev.id, active: !ev.is_active })}
                    className={[
                      "px-3 py-1.5 rounded-lg text-xs font-semibold transition",
                      ev.is_active
                        ? "bg-emerald-100 text-emerald-700 hover:bg-emerald-200"
                        : "bg-gray-100 text-gray-500 hover:bg-gray-200",
                    ].join(" ")}
                  >
                    {ev.is_active ? "● Активна" : "○ Выключена"}
                  </button>
                  <button
                    onClick={() => {
                      if (confirm(`Удалить акцию «${ev.name}»?`)) del.mutate(ev.id);
                    }}
                    className="px-2.5 py-1.5 rounded-lg text-xs font-semibold bg-red-50 text-red-600 hover:bg-red-100 transition"
                  >
                    Удалить
                  </button>
                </div>
              </div>
              {openId === ev.id && <SaleItemsEditor eventId={ev.id} locationId={locationId} />}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function SaleItemsEditor({ eventId, locationId }: { eventId: string; locationId: string }) {
  const qc = useQueryClient();
  const [pickId, setPickId] = useState("");
  const [price, setPrice] = useState("");

  const { data: detail } = useQuery<SaleDetail>({
    queryKey: ["sale", eventId],
    queryFn: () => api.get(`/sales/${eventId}`).then((r) => r.data.data),
  });
  const { data: products = [] } = useQuery<ProductDTO[]>({
    queryKey: ["products", locationId],
    queryFn: () => api.get("/products", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const items = detail?.items ?? [];
  const usedIds = new Set(items.map((i) => i.product_id));
  const available = products.filter((p) => !usedIds.has(p.id));
  const selected = products.find((p) => p.id === pickId);

  function invalidate() {
    qc.invalidateQueries({ queryKey: ["sale", eventId] });
    qc.invalidateQueries({ queryKey: ["sales", locationId] });
    qc.invalidateQueries({ queryKey: ["products"] });
  }

  const addItem = useMutation({
    mutationFn: () => api.post(`/sales/${eventId}/items`, { product_id: pickId, sale_price: Number(price) }),
    onSuccess: () => {
      setPickId("");
      setPrice("");
      invalidate();
    },
  });
  const removeItem = useMutation({
    mutationFn: (productId: string) => api.delete(`/sales/${eventId}/items/${productId}`),
    onSuccess: invalidate,
  });

  return (
    <div className="border-t border-gray-100 px-5 py-4 bg-gray-50/50 space-y-3">
      {items.length === 0 ? (
        <p className="text-sm text-gray-400">Добавьте товары и цену по акции.</p>
      ) : (
        <div className="space-y-1.5">
          {items.map((it) => (
            <div key={it.product_id} className="flex items-center gap-2 text-sm">
              <span className="flex-1 text-gray-800">{it.product_name}</span>
              <span className="text-gray-400 line-through text-xs">{it.base_price.toLocaleString("ru-RU")}</span>
              <span className="font-semibold text-amber-600">{it.sale_price.toLocaleString("ru-RU")} ₸</span>
              <button
                onClick={() => removeItem.mutate(it.product_id)}
                className="text-red-400 hover:text-red-600 text-lg leading-none px-1"
              >
                ×
              </button>
            </div>
          ))}
        </div>
      )}

      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (pickId && price && Number(price) >= 0) addItem.mutate();
        }}
        className="flex gap-2 pt-1"
      >
        <select
          value={pickId}
          onChange={(e) => setPickId(e.target.value)}
          className="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
        >
          <option value="">Товар…</option>
          {available.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name} ({p.base_price.toLocaleString("ru-RU")} ₸)
            </option>
          ))}
        </select>
        <input
          type="number"
          min="0"
          step="any"
          value={price}
          onChange={(e) => setPrice(e.target.value)}
          placeholder={selected ? `< ${selected.base_price}` : "Цена"}
          className="w-28 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
        />
        <button
          type="submit"
          disabled={addItem.isPending}
          className="px-3 py-2 rounded-lg bg-brand text-white text-sm font-semibold hover:bg-brand-600 transition disabled:opacity-50"
        >
          +
        </button>
      </form>
    </div>
  );
}
