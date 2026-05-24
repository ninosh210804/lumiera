import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";
import RecipeEditor from "../components/RecipeEditor";

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
  recipe_id: string | null;
}

export default function MenuPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const canEdit = user?.role === "admin" || user?.role === "manager";
  const qc = useQueryClient();
  const [activeCat, setActiveCat] = useState<string | null>(null);
  const [editing, setEditing] = useState<ProductDTO | "new" | null>(null);
  const [catModalOpen, setCatModalOpen] = useState(false);
  const [recipeFor, setRecipeFor] = useState<ProductDTO | null>(null);

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

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/products/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["products", locationId] }),
  });

  const filtered = activeCat
    ? products.filter((p) => p.category_id === activeCat)
    : products;

  const catName = (id: string) => categories.find((c) => c.id === id)?.name ?? "";

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Меню</h1>
        <div className="flex items-center gap-3">
          <span className="text-sm text-gray-400">{products.length} товаров</span>
          {canEdit && (
            <>
              <button
                onClick={() => setCatModalOpen(true)}
                className="text-sm font-semibold px-3 py-2 rounded-lg bg-gray-100 text-gray-700 hover:bg-gray-200 transition"
              >
                + Категория
              </button>
              <button
                onClick={() => setEditing("new")}
                className="text-sm font-semibold px-3 py-2 rounded-lg bg-brand text-white hover:bg-brand-600 transition"
              >
                + Товар
              </button>
            </>
          )}
        </div>
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
              {canEdit && (
                <button
                  onClick={() => setEditing("new")}
                  className="mt-2 text-sm font-semibold text-brand hover:underline"
                >
                  Добавить первый товар
                </button>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-3 gap-4">
              {filtered.map((product) => (
                <ProductCard
                  key={product.id}
                  product={product}
                  canEdit={canEdit}
                  onToggleStop={() =>
                    stopListMutation.mutate({ id: product.id, stop: !product.is_stop_listed })
                  }
                  onEdit={() => setEditing(product)}
                  onRecipe={() => setRecipeFor(product)}
                />
              ))}
            </div>
          )}
        </div>
      </div>

      {editing && (
        <ProductModal
          product={editing === "new" ? null : editing}
          categories={categories}
          onClose={() => setEditing(null)}
          onSaved={() => {
            qc.invalidateQueries({ queryKey: ["products", locationId] });
            setEditing(null);
          }}
          onDelete={
            editing === "new"
              ? undefined
              : () => {
                  if (confirm(`Удалить «${(editing as ProductDTO).name}»?`)) {
                    deleteMutation.mutate((editing as ProductDTO).id);
                    setEditing(null);
                  }
                }
          }
        />
      )}

      {catModalOpen && (
        <CategoryModal
          onClose={() => setCatModalOpen(false)}
          onSaved={() => {
            qc.invalidateQueries({ queryKey: ["categories", locationId] });
            setCatModalOpen(false);
          }}
        />
      )}

      {recipeFor && (
        <RecipeEditor
          product={recipeFor}
          locationId={locationId}
          categoryName={catName(recipeFor.category_id)}
          onClose={() => setRecipeFor(null)}
        />
      )}
    </div>
  );
}

function ProductCard({
  product,
  canEdit,
  onToggleStop,
  onEdit,
  onRecipe,
}: {
  product: ProductDTO;
  canEdit: boolean;
  onToggleStop: () => void;
  onEdit: () => void;
  onRecipe: () => void;
}) {
  return (
    <div
      className={[
        "bg-white rounded-2xl border p-4 flex flex-col gap-2 transition-all",
        product.is_stop_listed ? "border-red-100 opacity-60" : "border-gray-100",
        !product.is_active ? "grayscale" : "",
      ].join(" ")}
    >
      <div className="flex items-start justify-between gap-1">
        <p className="font-semibold text-gray-900 text-sm leading-tight">{product.name}</p>
        <div className="flex items-center gap-1 shrink-0">
          {!product.is_active && <span className="text-[10px] text-gray-400">скрыт</span>}
          {product.is_stop_listed && <span className="text-base leading-none">⛔</span>}
        </div>
      </div>

      {product.description && (
        <p className="text-xs text-gray-400 leading-tight line-clamp-2">{product.description}</p>
      )}

      <div className="mt-auto flex items-center justify-between pt-2">
        <span className="text-lg font-bold text-brand tabular-nums">
          {product.base_price.toLocaleString("ru-RU")} ₸
        </span>
        <div className="flex items-center gap-1.5">
          {canEdit && (
            <>
              <button
                onClick={onRecipe}
                title="Состав / ингредиенты"
                className="text-[11px] font-semibold px-2 py-1 rounded-lg bg-amber-50 text-amber-700 hover:bg-amber-100 transition"
              >
                🧾 Состав
              </button>
              <button
                onClick={onEdit}
                title="Редактировать"
                className="text-[11px] font-semibold px-2 py-1 rounded-lg bg-gray-100 text-gray-600 hover:bg-gray-200 transition"
              >
                ✏️
              </button>
            </>
          )}
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
    </div>
  );
}

function Modal({ title, children, onClose }: { title: string; children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div
        className="bg-white rounded-2xl shadow-xl w-full max-w-md max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <h2 className="text-lg font-bold text-gray-900">{title}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl leading-none">×</button>
        </div>
        <div className="p-6">{children}</div>
      </div>
    </div>
  );
}

const fieldClass =
  "w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand/30 focus:border-brand transition";
const labelClass = "block text-sm font-medium text-gray-700 mb-1";

function ProductModal({
  product,
  categories,
  onClose,
  onSaved,
  onDelete,
}: {
  product: ProductDTO | null;
  categories: CategoryDTO[];
  onClose: () => void;
  onSaved: () => void;
  onDelete?: () => void;
}) {
  const [name, setName] = useState(product?.name ?? "");
  const [categoryId, setCategoryId] = useState(product?.category_id ?? categories[0]?.id ?? "");
  const [price, setPrice] = useState(String(product?.base_price ?? ""));
  const [description, setDescription] = useState(product?.description ?? "");
  const [isActive, setIsActive] = useState(product?.is_active ?? true);
  const [error, setError] = useState("");

  const save = useMutation({
    mutationFn: () => {
      const body = {
        category_id: categoryId,
        name: name.trim(),
        description: description.trim(),
        base_price: Number(price),
        is_active: isActive,
        sort_order: product?.sort_order ?? 0,
      };
      return product ? api.put(`/products/${product.id}`, body) : api.post("/products", body);
    },
    onSuccess: onSaved,
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      setError(msg ?? "Не удалось сохранить");
    },
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!name.trim()) return setError("Введите название");
    if (!categoryId) return setError("Выберите категорию");
    if (!price || Number(price) < 0 || Number.isNaN(Number(price)))
      return setError("Введите корректную цену");
    save.mutate();
  }

  return (
    <Modal title={product ? "Редактировать товар" : "Новый товар"} onClose={onClose}>
      <form onSubmit={submit} className="space-y-4">
        <div>
          <label className={labelClass}>Название</label>
          <input className={fieldClass} value={name} onChange={(e) => setName(e.target.value)} placeholder="Капучино" autoFocus />
        </div>
        <div>
          <label className={labelClass}>Категория</label>
          <select className={fieldClass} value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
            {categories.length === 0 && <option value="">— сначала создайте категорию —</option>}
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.icon} {c.name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className={labelClass}>Цена, ₸</label>
          <input
            className={fieldClass}
            type="number"
            min="0"
            step="50"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            placeholder="1200"
          />
        </div>
        <div>
          <label className={labelClass}>Описание (опционально)</label>
          <textarea
            className={fieldClass}
            rows={2}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Эспрессо со вспененным молоком"
          />
        </div>
        <label className="flex items-center gap-2 text-sm text-gray-700">
          <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} className="rounded" />
          Показывать в меню
        </label>

        {error && <p className="text-red-500 text-sm bg-red-50 px-3 py-2 rounded-lg">{error}</p>}

        <div className="flex items-center gap-2 pt-2">
          <button
            type="submit"
            disabled={save.isPending}
            className="flex-1 bg-brand text-white py-2.5 rounded-lg text-sm font-semibold hover:bg-brand-600 transition disabled:opacity-50"
          >
            {save.isPending ? "Сохранение..." : "Сохранить"}
          </button>
          {onDelete && (
            <button
              type="button"
              onClick={onDelete}
              className="px-4 py-2.5 rounded-lg text-sm font-semibold bg-red-50 text-red-600 hover:bg-red-100 transition"
            >
              Удалить
            </button>
          )}
        </div>
      </form>
    </Modal>
  );
}

const EMOJI = ["☕", "🍵", "🥐", "🧊", "🍰", "🥤", "🍩", "🥪", "🍫", "🧃"];

function CategoryModal({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const [name, setName] = useState("");
  const [icon, setIcon] = useState(EMOJI[0]);
  const [error, setError] = useState("");

  const save = useMutation({
    mutationFn: () => api.post("/categories", { name: name.trim(), icon, sort_order: 0 }),
    onSuccess: onSaved,
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      setError(msg ?? "Не удалось сохранить");
    },
  });

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (!name.trim()) return setError("Введите название");
    save.mutate();
  }

  return (
    <Modal title="Новая категория" onClose={onClose}>
      <form onSubmit={submit} className="space-y-4">
        <div>
          <label className={labelClass}>Название</label>
          <input className={fieldClass} value={name} onChange={(e) => setName(e.target.value)} placeholder="Десерты" autoFocus />
        </div>
        <div>
          <label className={labelClass}>Иконка</label>
          <div className="flex flex-wrap gap-2">
            {EMOJI.map((e) => (
              <button
                type="button"
                key={e}
                onClick={() => setIcon(e)}
                className={[
                  "w-10 h-10 rounded-lg text-lg transition",
                  icon === e ? "bg-brand/10 ring-2 ring-brand" : "bg-gray-100 hover:bg-gray-200",
                ].join(" ")}
              >
                {e}
              </button>
            ))}
          </div>
        </div>

        {error && <p className="text-red-500 text-sm bg-red-50 px-3 py-2 rounded-lg">{error}</p>}

        <button
          type="submit"
          disabled={save.isPending}
          className="w-full bg-brand text-white py-2.5 rounded-lg text-sm font-semibold hover:bg-brand-600 transition disabled:opacity-50"
        >
          {save.isPending ? "Сохранение..." : "Создать"}
        </button>
      </form>
    </Modal>
  );
}
