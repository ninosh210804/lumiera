import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

interface QuoteDTO {
  subtotal: number;
  promo_discount: number;
  manual_discount: number;
  free_coffee_discount: number;
  free_coffees_applied: number;
  discount_total: number;
  loyalty_points_used: number;
  total: number;
  promo_active: boolean;
  promo_percent: number;
  customer_found: boolean;
  points_balance: number;
  free_drinks_left: number;
}

interface ProductDTO {
  id: string;
  name: string;
  category_name: string;
  base_price: number;
  sale_price: number | null;
  is_stop_listed: boolean;
  is_active: boolean;
}

interface PaymentMethodDTO {
  id: string;
  code: string;
  name: string;
}

interface OrderItem {
  product_id: string;
  product_name: string;
  qty: number;
  unit_price: number;
  line_total: number;
}

interface CreateOrderRequest {
  items: Array<{
    product_id: string;
    qty: number;
    modifier_option_ids: string[];
  }>;
  payments: Array<{
    method: string;
    amount: number;
    external_ref: string;
  }>;
  customer_phone: string;
  loyalty_points_to_use: number;
  manual_discount_pct: number;
  client_uuid: string;
  shift_id: string;
  comp?: boolean;
  comp_recipient?: string;
}

// Fixed discount applied when a customer presents a paper "-10%" ticket.
const TICKET_DISCOUNT_PCT = 10;

// Who takes products off the shelf without paying. Tracked per recipient in
// Analytics. Extend this list if more people start taking stock on credit.
const COMP_RECIPIENT = "Жандос";

export default function POSPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const qc = useQueryClient();
  const [cart, setCart] = useState<OrderItem[]>([]);
  const [activeCat, setActiveCat] = useState<string | null>(null);
  const [customerPhone, setCustomerPhone] = useState("");
  const [useBonuses, setUseBonuses] = useState(false);
  const [bonusAmount, setBonusAmount] = useState(0);
  const [ticketDiscount, setTicketDiscount] = useState(false);
  const [showOpenShift, setShowOpenShift] = useState(false);
  const [openingCash, setOpeningCash] = useState("");
  const [openShiftError, setOpenShiftError] = useState("");

  const { data: products = [] } = useQuery<ProductDTO[]>({
    queryKey: ["products-pos", locationId],
    queryFn: () =>
      api.get("/products", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: paymentMethods = [] } = useQuery<PaymentMethodDTO[]>({
    queryKey: ["payment-methods"],
    queryFn: () => api.get("/orders/payment-methods").then((r) => r.data.data ?? []),
  });

  const { data: activeShift, isLoading: shiftLoading } = useQuery({
    queryKey: ["active-shift"],
    queryFn: () =>
      api
        .get("/shifts/active")
        .then((r) => r.data.data)
        .catch(() => null),
  });

  const openShiftMutation = useMutation({
    mutationFn: (opening_cash: number) =>
      api.post("/shifts/open", {
        opening_cash,
        client_uuid: crypto.randomUUID(),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["active-shift"] });
      setShowOpenShift(false);
      setOpeningCash("");
      setOpenShiftError("");
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        (err instanceof Error ? err.message : "Не удалось открыть смену");
      setOpenShiftError(msg);
    },
  });

  const createOrderMutation = useMutation({
    mutationFn: (data: CreateOrderRequest) => api.post("/orders", data),
    onSuccess: (res) => {
      const earned = res.data?.data?.loyalty_earned ?? 0;
      setCart([]);
      setCustomerPhone("");
      setUseBonuses(false);
      setBonusAmount(0);
      setTicketDiscount(false);
      qc.invalidateQueries({ queryKey: ["active-shift"] });
      alert(
        earned > 0
          ? `Заказ создан успешно! Начислено ${earned.toLocaleString("ru-RU")} бонусов`
          : "Заказ создан успешно!"
      );
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        (err instanceof Error ? err.message : "Не удалось создать заказ");
      alert(`Ошибка: ${msg}`);
    },
  });

  // Debounce phone so we don't fire a quote on every keystroke.
  const [debouncedPhone, setDebouncedPhone] = useState("");
  useEffect(() => {
    const t = setTimeout(() => setDebouncedPhone(customerPhone), 350);
    return () => clearTimeout(t);
  }, [customerPhone]);

  useEffect(() => {
    setUseBonuses(false);
    setBonusAmount(0);
  }, [debouncedPhone]);

  const subtotal = cart.reduce((s, item) => s + item.line_total, 0);

  // Distinct categories of available products, for the menu filter.
  const categories = Array.from(
    new Set(
      products
        .filter((p) => !p.is_stop_listed && p.is_active)
        .map((p) => p.category_name)
        .filter(Boolean)
    )
  );

  // Authoritative price breakdown from the server (applies promo + loyalty).
  const quoteItems = cart.map((i) => ({
    product_id: i.product_id,
    qty: i.qty,
    modifier_option_ids: [] as string[],
  }));
  const manualDiscountPct = ticketDiscount ? TICKET_DISCOUNT_PCT : 0;
  const { data: quote } = useQuery<QuoteDTO>({
    queryKey: ["order-quote", quoteItems, debouncedPhone, bonusAmount, manualDiscountPct],
    queryFn: () =>
      api
        .post("/orders/quote", {
          items: quoteItems,
          customer_phone: debouncedPhone,
          loyalty_points_to_use: bonusAmount,
          manual_discount_pct: manualDiscountPct,
        })
        .then((r) => r.data.data),
    enabled: cart.length > 0 || debouncedPhone.length >= 4,
  });

  useEffect(() => {
    if (!useBonuses) {
      setBonusAmount(0);
      return;
    }
    if (!quote?.customer_found || (quote.points_balance ?? 0) <= 0) {
      setBonusAmount(0);
      return;
    }
    setBonusAmount(Math.min(quote.points_balance, Math.max(quote.total, 0)));
  }, [quote?.customer_found, quote?.points_balance, quote?.total, useBonuses]);

  const total = quote?.total ?? subtotal;

  const addToCart = (product: ProductDTO) => {
    if (product.is_stop_listed || !product.is_active) {
      alert("Этот товар недоступен");
      return;
    }
    const price = product.sale_price ?? product.base_price;
    setCart([
      ...cart,
      {
        product_id: product.id,
        product_name: product.name,
        qty: 1,
        unit_price: price,
        line_total: price,
      },
    ]);
  };

  const removeFromCart = (index: number) => {
    setCart(cart.filter((_, i) => i !== index));
  };

  const updateQty = (index: number, qty: number) => {
    if (qty < 1) {
      removeFromCart(index);
      return;
    }
    const updated = [...cart];
    updated[index].qty = qty;
    updated[index].line_total = updated[index].unit_price * qty;
    setCart(updated);
  };

  const handleSubmitOrder = (methodCode: string) => {
    if (cart.length === 0) {
      alert("Корзина пуста");
      return;
    }
    if (!activeShift) {
      alert("Смена не открыта");
      return;
    }
    if (total <= 0) {
      alert("Сумма заказа должна быть больше 0");
      return;
    }

    const request: CreateOrderRequest = {
      items: cart.map((item) => ({
        product_id: item.product_id,
        qty: item.qty,
        modifier_option_ids: [],
      })),
      payments: [
        {
          method: methodCode,
          amount: total,
          external_ref: "",
        },
      ],
      customer_phone: customerPhone,
      loyalty_points_to_use: useBonuses ? bonusAmount : 0,
      manual_discount_pct: manualDiscountPct,
      client_uuid: crypto.randomUUID(),
      shift_id: activeShift?.id ?? "",
    };

    createOrderMutation.mutate(request);
  };

  // Comp: products taken without payment. No payment is sent; the backend still
  // deducts ingredients from the warehouse and records it against the recipient.
  const handleCompOrder = () => {
    if (cart.length === 0) {
      alert("Корзина пуста");
      return;
    }
    if (!activeShift) {
      alert("Смена не открыта");
      return;
    }
    if (
      !confirm(
        `${COMP_RECIPIENT} забирает товар на ${subtotal.toLocaleString("ru-RU")} ₸ без оплаты?`
      )
    ) {
      return;
    }

    const request: CreateOrderRequest = {
      items: cart.map((item) => ({
        product_id: item.product_id,
        qty: item.qty,
        modifier_option_ids: [],
      })),
      payments: [],
      customer_phone: "",
      loyalty_points_to_use: 0,
      manual_discount_pct: 0,
      client_uuid: crypto.randomUUID(),
      shift_id: activeShift?.id ?? "",
      comp: true,
      comp_recipient: COMP_RECIPIENT,
    };

    createOrderMutation.mutate(request);
  };

  const handleOpenShift = () => {
    setOpenShiftError("");
    const cash = parseFloat(openingCash || "0");
    if (Number.isNaN(cash) || cash < 0) {
      setOpenShiftError("Введите корректную сумму");
      return;
    }
    openShiftMutation.mutate(cash);
  };

  return (
    <div className="h-screen bg-gray-900 flex flex-col">
      {/* Shift banner */}
      <div className="bg-gray-800 border-b border-gray-700 px-4 py-2 flex items-center justify-between">
        {shiftLoading ? (
          <span className="text-gray-400 text-sm">Загрузка смены...</span>
        ) : activeShift ? (
          <>
            <div className="text-sm text-gray-300">
              <span className="text-emerald-400 font-semibold">● Смена открыта</span>
              <span className="text-gray-500 ml-3">
                Открытие: {activeShift.opening_cash?.toLocaleString("ru-RU") ?? 0} ₸
              </span>
              {typeof activeShift.orders_count === "number" && (
                <span className="text-gray-500 ml-3">
                  Чеков: {activeShift.orders_count} ·{" "}
                  {(activeShift.revenue ?? 0).toLocaleString("ru-RU")} ₸
                </span>
              )}
            </div>
            <span className="text-gray-500 text-xs">{user?.full_name}</span>
          </>
        ) : (
          <>
            <span className="text-amber-400 text-sm font-medium">
              ⚠ Смена не открыта — заказы недоступны
            </span>
            <button
              onClick={() => {
                setOpenShiftError("");
                setOpeningCash("");
                setShowOpenShift(true);
              }}
              className="bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-semibold px-3 py-1.5 rounded transition"
            >
              Открыть смену
            </button>
          </>
        )}
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Left: Menu/Products */}
        <div className="flex-1 overflow-y-auto p-4">
          <h1 className="text-2xl font-bold text-white mb-4">Меню</h1>

          {/* Category filter */}
          {categories.length > 1 && (
            <div className="flex flex-wrap gap-2 mb-4">
              <button
                onClick={() => setActiveCat(null)}
                className={[
                  "px-3 py-1.5 rounded-lg text-sm font-medium transition",
                  activeCat === null
                    ? "bg-emerald-600 text-white"
                    : "bg-gray-700 text-gray-300 hover:bg-gray-600",
                ].join(" ")}
              >
                Все
              </button>
              {categories.map((cat) => (
                <button
                  key={cat}
                  onClick={() => setActiveCat(cat)}
                  className={[
                    "px-3 py-1.5 rounded-lg text-sm font-medium transition",
                    activeCat === cat
                      ? "bg-emerald-600 text-white"
                      : "bg-gray-700 text-gray-300 hover:bg-gray-600",
                  ].join(" ")}
                >
                  {cat}
                </button>
              ))}
            </div>
          )}

          <div className="grid grid-cols-2 gap-2">
            {products
              .filter((p) => !p.is_stop_listed && p.is_active)
              .filter((p) => activeCat === null || p.category_name === activeCat)
              .map((p) => (
                <button
                  key={p.id}
                  onClick={() => addToCart(p)}
                  disabled={!activeShift}
                  className="bg-gray-700 hover:bg-gray-600 disabled:opacity-40 disabled:cursor-not-allowed text-white p-3 rounded-lg text-sm font-medium transition"
                >
                  <div className="text-left">
                    <div className="font-semibold truncate">{p.name}</div>
                    <div className="text-gray-300 text-xs">{p.category_name}</div>
                    {p.sale_price != null ? (
                      <div className="mt-1 flex items-baseline gap-1.5">
                        <span className="text-gray-500 text-xs line-through">
                          {p.base_price.toLocaleString("ru-RU")}
                        </span>
                        <span className="text-amber-400 font-bold">
                          {p.sale_price.toLocaleString("ru-RU")} ₸
                        </span>
                        <span className="text-[9px] bg-amber-500/20 text-amber-300 px-1 rounded">
                          АКЦИЯ
                        </span>
                      </div>
                    ) : (
                      <div className="text-emerald-400 font-bold mt-1">
                        {p.base_price.toLocaleString("ru-RU")} ₸
                      </div>
                    )}
                  </div>
                </button>
              ))}
          </div>
        </div>

        {/* Right: Cart & Checkout */}
        <div className="w-80 bg-gray-800 border-l border-gray-700 flex flex-col">
          {/* Cart Items */}
          <div className="flex-1 overflow-y-auto p-4 space-y-2">
            <h2 className="text-xl font-bold text-white mb-3">Заказ</h2>
            {cart.length === 0 ? (
              <div className="text-gray-500 text-center py-8">Корзина пуста</div>
            ) : (
              cart.map((item, idx) => (
                <div key={idx} className="bg-gray-700 p-2 rounded text-sm text-white">
                  <div className="flex justify-between items-start mb-1">
                    <div className="flex-1">
                      <div className="font-medium">{item.product_name}</div>
                      <div className="text-gray-300 text-xs">
                        {item.unit_price.toLocaleString("ru-RU")} ₸/шт
                      </div>
                    </div>
                    <button
                      onClick={() => removeFromCart(idx)}
                      className="text-red-400 hover:text-red-300 font-bold ml-2"
                    >
                      ✕
                    </button>
                  </div>
                  <div className="flex items-center gap-1 justify-between">
                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => updateQty(idx, item.qty - 1)}
                        className="bg-gray-600 hover:bg-gray-500 px-2 py-1 rounded text-xs"
                      >
                        −
                      </button>
                      <span className="w-8 text-center font-bold">{item.qty}</span>
                      <button
                        onClick={() => updateQty(idx, item.qty + 1)}
                        className="bg-gray-600 hover:bg-gray-500 px-2 py-1 rounded text-xs"
                      >
                        +
                      </button>
                    </div>
                    <div className="font-bold text-emerald-400">
                      {item.line_total.toLocaleString("ru-RU")} ₸
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Customer Phone */}
          <div className="border-t border-gray-700 p-4">
            <label className="block text-gray-300 text-xs font-medium mb-1">
              Телефон клиента (опционально)
            </label>
            <input
              type="tel"
              value={customerPhone}
              onChange={(e) => setCustomerPhone(e.target.value)}
              placeholder="+7 (700) 000-0000"
              className="w-full bg-gray-700 text-white px-3 py-2 rounded border border-gray-600 text-sm"
            />
            {quote?.customer_found && (
              <div className="mt-2 flex flex-col gap-2 text-xs">
                <div className="flex items-center gap-2">
                  <span className="bg-amber-500/15 text-amber-300 px-2 py-1 rounded-md font-medium">
                    ☕ Бесплатных: {quote.free_drinks_left}
                  </span>
                  <span className="bg-sky-500/15 text-sky-300 px-2 py-1 rounded-md font-medium">
                    💰 Бонусы: {quote.points_balance.toLocaleString("ru-RU")}
                  </span>
                </div>
                {quote.points_balance > 0 && (
                  <button
                    type="button"
                    onClick={() => {
                      if (!quote.customer_found || quote.points_balance <= 0) return;
                      setUseBonuses((prev) => !prev);
                    }}
                    className={`w-full rounded-md border px-2 py-1.5 text-left text-xs font-medium transition ${
                      useBonuses
                        ? "border-emerald-500/40 bg-emerald-500/15 text-emerald-300"
                        : "border-gray-600 bg-gray-700 text-gray-200 hover:bg-gray-600"
                    }`}
                  >
                    {useBonuses
                      ? `✅ Используются бонусы: ${bonusAmount.toLocaleString("ru-RU")} ₸`
                      : "Использовать бонусы"}
                  </button>
                )}
              </div>
            )}
            {/* Paper "-10%" ticket: per-order discount the barista toggles on. */}
            <button
              type="button"
              onClick={() => setTicketDiscount((prev) => !prev)}
              className={`mt-2 w-full rounded-md border px-2 py-1.5 text-left text-xs font-medium transition ${
                ticketDiscount
                  ? "border-emerald-500/40 bg-emerald-500/15 text-emerald-300"
                  : "border-gray-600 bg-gray-700 text-gray-200 hover:bg-gray-600"
              }`}
            >
              {ticketDiscount
                ? `✅ Скидка по талону −${TICKET_DISCOUNT_PCT}%`
                : `🎟 Скидка по талону −${TICKET_DISCOUNT_PCT}%`}
            </button>
          </div>

          {/* Totals */}
          <div className="border-t border-gray-700 p-4 space-y-2">
            {quote?.promo_active && (
              <div className="bg-emerald-500/15 text-emerald-300 text-xs font-semibold px-2.5 py-1.5 rounded-md text-center">
                🎉 Акция −{quote.promo_percent}% на всё
              </div>
            )}
            <div className="flex justify-between text-gray-300">
              <span>Товары:</span>
              <span>{subtotal.toLocaleString("ru-RU")} ₸</span>
            </div>
            {!!quote && quote.free_coffee_discount > 0 && (
              <div className="flex justify-between text-amber-300">
                <span>Бесплатный кофе ×{quote.free_coffees_applied}:</span>
                <span>−{quote.free_coffee_discount.toLocaleString("ru-RU")} ₸</span>
              </div>
            )}
            {!!quote && quote.promo_discount > 0 && (
              <div className="flex justify-between text-emerald-300">
                <span>Скидка −{quote.promo_percent}%:</span>
                <span>−{quote.promo_discount.toLocaleString("ru-RU")} ₸</span>
              </div>
            )}
            {!!quote && quote.manual_discount > 0 && (
              <div className="flex justify-between text-emerald-300">
                <span>Скидка по талону −{TICKET_DISCOUNT_PCT}%:</span>
                <span>−{quote.manual_discount.toLocaleString("ru-RU")} ₸</span>
              </div>
            )}
            {!!quote && quote.loyalty_points_used > 0 && (
              <div className="flex justify-between text-sky-300">
                <span>Бонусы к оплате:</span>
                <span>−{quote.loyalty_points_used.toLocaleString("ru-RU")} ₸</span>
              </div>
            )}
            <div className="border-t border-gray-600 pt-2 flex justify-between text-lg font-bold text-white">
              <span>ИТОГО:</span>
              <span className="text-emerald-400">{total.toLocaleString("ru-RU")} ₸</span>
            </div>
          </div>

          {/* Payment Buttons */}
          <div className="border-t border-gray-700 p-4 space-y-2">
            {paymentMethods.slice(0, 2).map((method) => (
              <button
                key={method.id}
                onClick={() => handleSubmitOrder(method.code)}
                disabled={createOrderMutation.isPending || cart.length === 0 || !activeShift}
                className="w-full bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-600 disabled:cursor-not-allowed text-white font-bold py-2 rounded transition"
              >
                {createOrderMutation.isPending ? "Обработка..." : `${method.name}`}
              </button>
            ))}
            <button
              onClick={handleCompOrder}
              disabled={createOrderMutation.isPending || cart.length === 0 || !activeShift}
              className="w-full bg-amber-600 hover:bg-amber-500 disabled:bg-gray-600 disabled:cursor-not-allowed text-white font-bold py-2 rounded transition"
            >
              {`${COMP_RECIPIENT} забрал (без оплаты)`}
            </button>
            <button
              onClick={() => setCart([])}
              className="w-full bg-gray-700 hover:bg-gray-600 text-white font-medium py-2 rounded transition"
            >
              Отмена
            </button>
          </div>
        </div>
      </div>

      {/* Open shift modal */}
      {showOpenShift && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
          <div className="bg-white rounded-2xl shadow-xl p-6 w-full max-w-sm mx-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold text-gray-900">Открыть смену</h3>
              <button
                onClick={() => {
                  setShowOpenShift(false);
                  setOpenShiftError("");
                }}
                className="text-gray-400 hover:text-gray-600 text-xl leading-none"
              >
                ×
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Начальная касса (₸)
                </label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  value={openingCash}
                  onChange={(e) => setOpeningCash(e.target.value)}
                  placeholder="0"
                  className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-400/30"
                  autoFocus
                />
                {openShiftError && <p className="text-xs text-red-500 mt-1">{openShiftError}</p>}
              </div>
              <div className="flex gap-2 justify-end">
                <button
                  onClick={() => {
                    setShowOpenShift(false);
                    setOpenShiftError("");
                  }}
                  className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700 transition"
                >
                  Отмена
                </button>
                <button
                  onClick={handleOpenShift}
                  disabled={openShiftMutation.isPending}
                  className="px-5 py-2 text-sm bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-semibold transition disabled:opacity-50"
                >
                  {openShiftMutation.isPending ? "Открытие..." : "Открыть"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
