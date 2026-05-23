import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

interface CategoryDTO {
  id: string;
  name: string;
  icon: string;
  sort_order: number;
  is_active: boolean;
}

interface ProductDTO {
  id: string;
  category_id: string;
  name: string;
  description: string | null;
  base_price: number;
  is_active: boolean;
  is_stop_listed: boolean;
  sort_order: number;
}

export default function MenuPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const qc = useQueryClient();
  const [activeCat, setActiveCat] = useState<string | null>(null);

  const { data: categories = [] } = useQuery<CategoryDTO[]>({
    queryKey: ["categories", locationId],
    queryFn: () =>
      api.get("/categories", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: products = [], isLoading } = useQuery<ProductDTO[]>({
    queryKey: ["products", locationId],
    queryFn: () =>
      api.get("/products", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const stopListMutation = useMutation({
    mutationFn: ({ id, stop }: { id: string; stop: boolean }) =>
      api.post(`/products/${id}/stop-list`, { stopped: stop }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["products", locationId] }),
  });

  const filtered = activeCat
    ? products.filter((p) => p.category_id === activeCat)
    : products;

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Меню</h1>
        <span className="text-sm text-gray-400">{products.length} товаров</span>
      </div>

      <div className="flex gap-6">
        {/* Category sidebar */}
        <div className="w-44 shrink-0 space-y-1">
          <button
            onClick={() => setActiveCat(null)}
            className={[
              "w-full flex items-center gap-2 px-3 py-2.5 rounded-xl text-sm font-medium transition-colors text-left",
              activeCat === null ? "bg-brand text-white" : "text-gray-500 hover:bg-gray-100",
            ].join(" ")}
          >
            <span>🍽️</span> Все категории
          </button>
          {categories.map((cat) => (
            <button
              key={cat.id}
              onClick={() => setActiveCat(cat.id)}
              className={[
                "w-full flex items-center gap-2 px-3 py-2.5 rounded-xl text-sm font-medium transition-colors text-left",
                activeCat === cat.id ? "bg-brand text-white" : "text-gray-500 hover:bg-gray-100",
              ].join(" ")}
            >
              <span>{cat.icon}</span>
              <span className="truncate">{cat.name}</span>
            </button>
          ))}
        </div>

        {/* Products grid */}
        <div className="flex-1">
          {isLoading ? (
            <div className="grid grid-cols-3 gap-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="h-36 bg-gray-100 rounded-2xl animate-pulse" />
              ))}
            </div>
          ) : filtered.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-64 text-gray-300 gap-2">
              <span className="text-4xl">☕</span>
              <p className="text-sm">Нет товаров</p>
            </div>
          ) : (
            <div className="grid grid-cols-3 gap-4">
              {filtered.map((product) => (
                <ProductCard
                  key={product.id}
                  product={product}
                  onToggleStop={() =>
                    stopListMutation.mutate({ id: product.id, stop: !product.is_stop_listed })
                  }
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function ProductCard({
  product,
  onToggleStop,
}: {
  product: ProductDTO;
  onToggleStop: () => void;
}) {
  return (
    <div
      className={[
        "bg-white rounded-2xl border p-4 flex flex-col gap-2 transition-all",
        product.is_stop_listed ? "border-red-100 opacity-60" : "border-gray-100",
      ].join(" ")}
    >
      <div className="flex items-start justify-between gap-1">
        <p className="font-semibold text-gray-900 text-sm leading-tight">{product.name}</p>
        {product.is_stop_listed && (
          <span className="text-base leading-none shrink-0">⛔</span>
        )}
      </div>

      {product.description && (
        <p className="text-xs text-gray-400 leading-tight line-clamp-2">{product.description}</p>
      )}

      <div className="mt-auto flex items-center justify-between pt-2">
        <span className="text-lg font-bold text-brand tabular-nums">
          {product.base_price.toLocaleString("ru-RU")} ₸
        </span>
        <button
          onClick={onToggleStop}
          className={[
            "text-[11px] font-semibold px-2.5 py-1 rounded-lg transition-colors",
            product.is_stop_listed
              ? "bg-emerald-50 text-emerald-700 hover:bg-emerald-100"
              : "bg-red-50 text-red-600 hover:bg-red-100",
          ].join(" ")}
        >
          {product.is_stop_listed ? "Вернуть" : "Стоп"}
        </button>
      </div>
    </div>
  );
}
