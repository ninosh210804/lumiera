import { useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "../lib/api";

interface MenuProductDTO {
  id: string;
  name: string;
  category_name: string;
  base_price: number;
  sale_price: number | null;
}

interface ClientProfileDTO {
  phone: string;
  customer_found: boolean;
  balance: number;
  free_drinks_left: number;
}

interface OnlineOrderItemDTO {
  product_id: string;
  product_name: string;
  qty: number;
  unit_price: number;
  line_total: number;
}

interface OnlineOrderDTO {
  id: string;
  status: "new" | "preparing" | "ready" | "completed" | "rejected";
  delivery_office: string;
  total: number;
  items: OnlineOrderItemDTO[];
}

const TRACK_KEY = "delivery_order_id";

const STATUS_STEPS: Array<{ key: string; label: string }> = [
  { key: "new", label: "Принят" },
  { key: "preparing", label: "Готовится" },
  { key: "ready", label: "Готов" },
];

const price = (p: MenuProductDTO) => p.sale_price ?? p.base_price;

export default function DeliveryPage() {
  const [params] = useSearchParams();
  const locationId = params.get("location") ?? "";

  const [phone, setPhone] = useState("");
  const [signedIn, setSignedIn] = useState(false);
  const [cart, setCart] = useState<Record<string, { product: MenuProductDTO; qty: number }>>({});
  const [office, setOffice] = useState("");
  const [note, setNote] = useState("");
  const [trackingId, setTrackingId] = useState<string | null>(() =>
    localStorage.getItem(TRACK_KEY)
  );

  const { data: menu = [] } = useQuery<MenuProductDTO[]>({
    queryKey: ["delivery-menu", locationId],
    queryFn: () =>
      api.get("/menu", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: Boolean(locationId),
  });

  const { data: profile } = useQuery<ClientProfileDTO>({
    queryKey: ["delivery-profile", phone],
    queryFn: () => api.get("/clients/profile", { params: { phone } }).then((r) => r.data.data),
    enabled: signedIn && phone.length >= 4,
    retry: false,
  });

  const { data: tracked } = useQuery<OnlineOrderDTO>({
    queryKey: ["delivery-track", trackingId],
    queryFn: () => api.get(`/clients/orders/${trackingId}`).then((r) => r.data.data),
    enabled: Boolean(trackingId),
    refetchInterval: (query) => {
      const s = query.state.data?.status;
      return s === "completed" || s === "rejected" ? false : 8000;
    },
  });

  const placeOrder = useMutation({
    mutationFn: () =>
      api
        .post("/clients/orders", {
          location_id: locationId,
          customer_phone: phone,
          delivery_office: office,
          delivery_note: note,
          items: Object.values(cart).map((c) => ({
            product_id: c.product.id,
            qty: c.qty,
            modifier_option_ids: [],
          })),
        })
        .then((r) => r.data.data as OnlineOrderDTO),
    onSuccess: (order) => {
      localStorage.setItem(TRACK_KEY, order.id);
      setTrackingId(order.id);
      setCart({});
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        "Не удалось оформить заказ";
      alert(msg);
    },
  });

  const cartLines = Object.values(cart);
  const subtotal = cartLines.reduce((s, c) => s + price(c.product) * c.qty, 0);

  const categories = useMemo(
    () => Array.from(new Set(menu.map((p) => p.category_name).filter(Boolean))),
    [menu]
  );

  const setQty = (product: MenuProductDTO, delta: number) => {
    setCart((prev) => {
      const cur = prev[product.id]?.qty ?? 0;
      const next = cur + delta;
      const copy = { ...prev };
      if (next <= 0) delete copy[product.id];
      else copy[product.id] = { product, qty: next };
      return copy;
    });
  };

  const startNewOrder = () => {
    localStorage.removeItem(TRACK_KEY);
    setTrackingId(null);
  };

  // ── No location in the QR link ───────────────────────────────────────────
  if (!locationId) {
    return (
      <div className="min-h-screen bg-gray-950 text-white flex items-center justify-center p-6">
        <div className="max-w-md rounded-2xl border border-gray-800 bg-gray-900 p-6 text-center">
          <p className="text-lg font-semibold">Ссылка недействительна</p>
          <p className="mt-2 text-sm text-gray-400">
            Отсканируйте QR-код заведения ещё раз, чтобы открыть меню.
          </p>
        </div>
      </div>
    );
  }

  // ── Tracking view ────────────────────────────────────────────────────────
  if (trackingId && tracked) {
    const stepIdx = STATUS_STEPS.findIndex((s) => s.key === tracked.status);
    const rejected = tracked.status === "rejected";
    const completed = tracked.status === "completed";
    const ready = tracked.status === "ready";
    return (
      <div className="min-h-screen bg-gray-950 text-white">
        <div className="mx-auto max-w-lg p-6">
          <h1 className="text-2xl font-bold">Ваш заказ</h1>
          <p className="mt-1 text-sm text-gray-400">Офис: {tracked.delivery_office || "—"}</p>

          {ready && (
            <div className="mt-6 rounded-2xl border border-emerald-500/40 bg-emerald-500/15 p-6 text-center">
              <p className="text-3xl font-bold text-emerald-300">✅ Заказ готов!</p>
              <p className="mt-2 text-sm text-emerald-200">Курьер уже несёт его вам.</p>
            </div>
          )}
          {completed && (
            <div className="mt-6 rounded-2xl border border-sky-500/40 bg-sky-500/15 p-6 text-center">
              <p className="text-2xl font-bold text-sky-300">Заказ выдан. Спасибо! ☕</p>
            </div>
          )}
          {rejected && (
            <div className="mt-6 rounded-2xl border border-red-500/40 bg-red-500/15 p-6 text-center">
              <p className="text-2xl font-bold text-red-300">Заказ отклонён</p>
              <p className="mt-2 text-sm text-red-200">Свяжитесь с заведением для уточнения.</p>
            </div>
          )}

          {!rejected && !completed && (
            <div className="mt-8 flex items-center justify-between">
              {STATUS_STEPS.map((step, i) => (
                <div key={step.key} className="flex flex-1 flex-col items-center">
                  <div
                    className={`flex h-10 w-10 items-center justify-center rounded-full text-sm font-bold ${
                      i <= stepIdx
                        ? "bg-emerald-500 text-white"
                        : "bg-gray-800 text-gray-500 border border-gray-700"
                    }`}
                  >
                    {i + 1}
                  </div>
                  <span
                    className={`mt-2 text-xs ${i <= stepIdx ? "text-emerald-300" : "text-gray-500"}`}
                  >
                    {step.label}
                  </span>
                </div>
              ))}
            </div>
          )}

          <div className="mt-8 rounded-2xl border border-gray-800 bg-gray-900 p-4">
            {tracked.items.map((it) => (
              <div key={it.product_id} className="flex justify-between py-1 text-sm">
                <span className="text-gray-300">
                  {it.product_name} × {it.qty}
                </span>
                <span className="text-gray-400">{it.line_total.toLocaleString("ru-RU")} ₸</span>
              </div>
            ))}
            <div className="mt-2 flex justify-between border-t border-gray-800 pt-2 font-bold">
              <span>Итого</span>
              <span className="text-emerald-400">{tracked.total.toLocaleString("ru-RU")} ₸</span>
            </div>
          </div>

          <button
            onClick={startNewOrder}
            className="mt-6 w-full rounded-xl bg-gray-800 py-3 text-sm font-semibold text-gray-200 hover:bg-gray-700"
          >
            Новый заказ
          </button>
        </div>
      </div>
    );
  }

  // ── Phone sign-in ────────────────────────────────────────────────────────
  if (!signedIn) {
    return (
      <div className="min-h-screen bg-gray-950 text-white flex items-center justify-center p-6">
        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (phone.trim().length >= 4) setSignedIn(true);
          }}
          className="w-full max-w-sm rounded-2xl border border-gray-800 bg-gray-900 p-6"
        >
          <p className="text-sm font-medium uppercase tracking-[0.25em] text-emerald-400">
            Доставка в офис
          </p>
          <h1 className="mt-2 text-2xl font-bold">Войдите по номеру</h1>
          <p className="mt-2 text-sm text-gray-400">На этот номер начисляются бонусы за заказы.</p>
          <input
            type="tel"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            placeholder="+7 700 000 00 00"
            className="mt-4 w-full rounded-xl border border-gray-700 bg-gray-800 px-4 py-3 text-white outline-none"
          />
          <button
            type="submit"
            className="mt-4 w-full rounded-xl bg-emerald-600 py-3 font-semibold text-white hover:bg-emerald-500"
          >
            Продолжить
          </button>
        </form>
      </div>
    );
  }

  // ── Shop: menu + cart ────────────────────────────────────────────────────
  return (
    <div className="min-h-screen bg-gray-950 text-white pb-40">
      <div className="mx-auto max-w-2xl p-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold">Меню</h1>
            <p className="text-xs text-gray-400">{phone}</p>
          </div>
          {profile?.customer_found && (
            <span className="rounded-md bg-sky-500/15 px-2 py-1 text-xs font-medium text-sky-300">
              💰 Бонусы: {profile.balance.toLocaleString("ru-RU")}
            </span>
          )}
        </div>

        {categories.map((cat) => (
          <div key={cat} className="mt-5">
            <h2 className="mb-2 text-sm font-semibold text-gray-400">{cat}</h2>
            <div className="space-y-2">
              {menu
                .filter((p) => p.category_name === cat)
                .map((p) => {
                  const qty = cart[p.id]?.qty ?? 0;
                  return (
                    <div
                      key={p.id}
                      className="flex items-center justify-between rounded-xl border border-gray-800 bg-gray-900 p-3"
                    >
                      <div>
                        <p className="font-medium">{p.name}</p>
                        <p className="text-sm text-emerald-400">
                          {price(p).toLocaleString("ru-RU")} ₸
                        </p>
                      </div>
                      {qty === 0 ? (
                        <button
                          onClick={() => setQty(p, 1)}
                          className="rounded-lg bg-emerald-600 px-4 py-2 text-sm font-semibold hover:bg-emerald-500"
                        >
                          +
                        </button>
                      ) : (
                        <div className="flex items-center gap-3">
                          <button
                            onClick={() => setQty(p, -1)}
                            className="rounded-lg bg-gray-700 px-3 py-1.5 font-bold hover:bg-gray-600"
                          >
                            −
                          </button>
                          <span className="w-5 text-center font-bold">{qty}</span>
                          <button
                            onClick={() => setQty(p, 1)}
                            className="rounded-lg bg-gray-700 px-3 py-1.5 font-bold hover:bg-gray-600"
                          >
                            +
                          </button>
                        </div>
                      )}
                    </div>
                  );
                })}
            </div>
          </div>
        ))}
      </div>

      {/* Sticky checkout bar */}
      {cartLines.length > 0 && (
        <div className="fixed inset-x-0 bottom-0 border-t border-gray-800 bg-gray-900/95 backdrop-blur">
          <div className="mx-auto max-w-2xl space-y-3 p-4">
            <div className="grid grid-cols-2 gap-2">
              <input
                value={office}
                onChange={(e) => setOffice(e.target.value)}
                placeholder="Офис / кабинет"
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm outline-none"
              />
              <input
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="Этаж, комментарий"
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm outline-none"
              />
            </div>
            <button
              onClick={() => {
                if (!office.trim()) {
                  alert("Укажите офис или кабинет для доставки");
                  return;
                }
                placeOrder.mutate();
              }}
              disabled={placeOrder.isPending}
              className="w-full rounded-xl bg-emerald-600 py-3 font-bold text-white hover:bg-emerald-500 disabled:opacity-50"
            >
              {placeOrder.isPending
                ? "Оформляем..."
                : `Оформить заказ · ${subtotal.toLocaleString("ru-RU")} ₸`}
            </button>
            <p className="text-center text-xs text-gray-500">Оплата при получении</p>
          </div>
        </div>
      )}
    </div>
  );
}
