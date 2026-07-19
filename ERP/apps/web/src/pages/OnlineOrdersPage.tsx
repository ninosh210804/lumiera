import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";

interface OnlineOrderItemDTO {
  product_id: string;
  product_name: string;
  qty: number;
  line_total: number;
}

interface OnlineOrderDTO {
  id: string;
  customer_phone: string;
  status: "new" | "preparing" | "ready" | "completed" | "rejected";
  delivery_office: string;
  delivery_note: string;
  total: number;
  created_at: string;
  items: OnlineOrderItemDTO[];
}

interface PaymentMethodDTO {
  id: string;
  code: string;
  name: string;
}

const STATUS_META: Record<string, { label: string; badge: string }> = {
  new: { label: "Новый", badge: "bg-amber-100 text-amber-700" },
  preparing: { label: "Готовится", badge: "bg-sky-100 text-sky-700" },
  ready: { label: "Готов", badge: "bg-emerald-100 text-emerald-700" },
};

export default function OnlineOrdersPage() {
  const qc = useQueryClient();
  const [payFor, setPayFor] = useState<OnlineOrderDTO | null>(null);

  const { data: orders = [] } = useQuery<OnlineOrderDTO[]>({
    queryKey: ["online-orders"],
    queryFn: () => api.get("/online-orders").then((r) => r.data.data ?? []),
    refetchInterval: 10000,
  });

  const { data: activeShift } = useQuery({
    queryKey: ["active-shift"],
    queryFn: () =>
      api
        .get("/shifts/active")
        .then((r) => r.data.data)
        .catch(() => null),
  });

  const { data: paymentMethods = [] } = useQuery<PaymentMethodDTO[]>({
    queryKey: ["payment-methods"],
    queryFn: () => api.get("/orders/payment-methods").then((r) => r.data.data ?? []),
  });

  const onError = (err: unknown) => {
    const msg =
      (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
      "Не удалось обновить заказ";
    alert(msg);
  };

  const setStatus = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      api.post(`/online-orders/${id}/status`, { status }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["online-orders"] }),
    onError,
  });

  const complete = useMutation({
    mutationFn: ({ id, method }: { id: string; method: string }) =>
      api.post(`/online-orders/${id}/complete`, {
        payment_method: method,
        shift_id: activeShift?.id ?? "",
      }),
    onSuccess: () => {
      setPayFor(null);
      qc.invalidateQueries({ queryKey: ["online-orders"] });
      qc.invalidateQueries({ queryKey: ["active-shift"] });
    },
    onError,
  });

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="mx-auto max-w-5xl">
        <h1 className="text-2xl font-bold text-gray-900">Онлайн-заказы</h1>
        <p className="mt-1 text-sm text-gray-500">
          Заказы на доставку в офис. Обновляется автоматически.
        </p>

        {!activeShift && (
          <div className="mt-4 rounded-lg border border-amber-200 bg-amber-50 px-4 py-2 text-sm text-amber-700">
            ⚠ Смена не открыта — принять оплату можно только при открытой смене.
          </div>
        )}

        {orders.length === 0 ? (
          <div className="mt-10 rounded-2xl border border-dashed border-gray-200 bg-white p-10 text-center text-gray-400">
            Активных заказов нет.
          </div>
        ) : (
          <div className="mt-6 grid gap-4 md:grid-cols-2">
            {orders.map((o) => {
              const meta = STATUS_META[o.status];
              return (
                <div
                  key={o.id}
                  className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm"
                >
                  <div className="flex items-start justify-between">
                    <div>
                      <p className="font-semibold text-gray-900">📍 {o.delivery_office || "—"}</p>
                      <p className="text-sm text-gray-500">{o.customer_phone}</p>
                      {o.delivery_note && (
                        <p className="mt-0.5 text-xs text-gray-400">{o.delivery_note}</p>
                      )}
                    </div>
                    <span
                      className={`rounded-full px-2.5 py-1 text-xs font-semibold ${meta?.badge ?? "bg-gray-100 text-gray-600"}`}
                    >
                      {meta?.label ?? o.status}
                    </span>
                  </div>

                  <div className="mt-3 space-y-1 border-t border-gray-100 pt-3">
                    {o.items.map((it) => (
                      <div key={it.product_id} className="flex justify-between text-sm">
                        <span className="text-gray-700">
                          {it.product_name} × {it.qty}
                        </span>
                        <span className="text-gray-500">
                          {it.line_total.toLocaleString("ru-RU")} ₸
                        </span>
                      </div>
                    ))}
                    <div className="flex justify-between border-t border-gray-100 pt-2 font-bold text-gray-900">
                      <span>Итого</span>
                      <span>{o.total.toLocaleString("ru-RU")} ₸</span>
                    </div>
                  </div>

                  <div className="mt-4 flex flex-wrap gap-2">
                    {o.status === "new" && (
                      <button
                        onClick={() => setStatus.mutate({ id: o.id, status: "preparing" })}
                        className="rounded-lg bg-sky-600 px-3 py-2 text-sm font-semibold text-white hover:bg-sky-500"
                      >
                        Принять
                      </button>
                    )}
                    {o.status === "preparing" && (
                      <button
                        onClick={() => setStatus.mutate({ id: o.id, status: "ready" })}
                        className="rounded-lg bg-emerald-600 px-3 py-2 text-sm font-semibold text-white hover:bg-emerald-500"
                      >
                        Готов
                      </button>
                    )}
                    {o.status === "ready" && (
                      <button
                        onClick={() => setPayFor(o)}
                        disabled={!activeShift}
                        className="rounded-lg bg-emerald-700 px-3 py-2 text-sm font-semibold text-white hover:bg-emerald-600 disabled:opacity-50"
                      >
                        Выдан · оплата
                      </button>
                    )}
                    {o.status !== "ready" && (
                      <button
                        onClick={() => {
                          if (confirm("Отклонить заказ?"))
                            setStatus.mutate({ id: o.id, status: "rejected" });
                        }}
                        className="rounded-lg bg-gray-100 px-3 py-2 text-sm font-medium text-gray-600 hover:bg-gray-200"
                      >
                        Отклонить
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Payment picker modal */}
      {payFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold text-gray-900">Оплата заказа</h3>
              <button
                onClick={() => setPayFor(null)}
                className="text-xl leading-none text-gray-400 hover:text-gray-600"
              >
                ×
              </button>
            </div>
            <p className="mt-1 text-sm text-gray-500">
              {payFor.delivery_office} · {payFor.total.toLocaleString("ru-RU")} ₸
            </p>
            <div className="mt-4 space-y-2">
              {paymentMethods
                .filter((m) => m.code !== "loyalty" && m.code !== "gift")
                .map((m) => (
                  <button
                    key={m.id}
                    onClick={() => complete.mutate({ id: payFor.id, method: m.code })}
                    disabled={complete.isPending}
                    className="w-full rounded-lg bg-emerald-600 py-2.5 font-semibold text-white hover:bg-emerald-500 disabled:opacity-50"
                  >
                    {m.name}
                  </button>
                ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
