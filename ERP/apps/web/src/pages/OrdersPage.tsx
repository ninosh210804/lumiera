import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { format, subDays, startOfDay } from "date-fns";
import { ru } from "date-fns/locale";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

interface OrderDTO {
  id: string;
  barista_id: string;
  status: string;
  subtotal: number;
  discount_total: number;
  total: number;
  receipt_no: string;
  created_at: string;
  items?: { product_name: string; qty: number; line_total: number }[];
  payments?: { method_name: string; amount: number }[];
}

type Range = "today" | "week" | "month";
const RANGES: { value: Range; label: string }[] = [
  { value: "today", label: "Сегодня" },
  { value: "week",  label: "Неделя" },
  { value: "month", label: "Месяц" },
];

const STATUS_LABELS: Record<string, { label: string; cls: string }> = {
  paid:      { label: "Оплачен",  cls: "bg-emerald-50 text-emerald-700" },
  pending:   { label: "Ожидание", cls: "bg-amber-50 text-amber-700" },
  cancelled: { label: "Отменён",  cls: "bg-red-50 text-red-600" },
};

function rangeDates(r: Range) {
  const to = new Date();
  const from =
    r === "today" ? startOfDay(to) :
    r === "week"  ? startOfDay(subDays(to, 6)) :
                    startOfDay(subDays(to, 29));
  return { from, to };
}

export default function OrdersPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const isAdmin = user?.role === "admin";
  const qc = useQueryClient();
  const [range, setRange] = useState<Range>("today");
  const [selected, setSelected] = useState<OrderDTO | null>(null);

  const refreshOrders = () => {
    qc.invalidateQueries({ queryKey: ["orders", locationId, range] });
    setSelected(null);
  };
  const softDelete = useMutation({
    mutationFn: (id: string) => api.post(`/orders/${id}/soft-delete`),
    onSuccess: refreshOrders,
  });
  const hardDelete = useMutation({
    mutationFn: (id: string) => api.delete(`/orders/${id}`),
    onSuccess: refreshOrders,
  });

  const { from, to } = useMemo(() => rangeDates(range), [range]);

  const { data: orders = [], isLoading } = useQuery<OrderDTO[]>({
    queryKey: ["orders", locationId, range],
    queryFn: () =>
      api
        .get("/orders", {
          params: {
            location_id: locationId,
            from: from.toISOString(),
            to: to.toISOString(),
          },
        })
        .then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: detail } = useQuery<OrderDTO>({
    queryKey: ["order", selected?.id],
    queryFn: () =>
      api.get(`/orders/${selected!.id}`).then((r) => r.data.data),
    enabled: !!selected,
  });

  const totalRevenue = orders.reduce((s, o) => s + (o.status === "paid" ? o.total : 0), 0);
  const paidCount = orders.filter((o) => o.status === "paid").length;

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Заказы</h1>
          <p className="text-sm text-gray-400 mt-0.5">
            {paidCount} чеков · {totalRevenue.toLocaleString("ru-RU")} ₸
          </p>
        </div>
        <div className="flex gap-1 bg-gray-100 p-1 rounded-xl">
          {RANGES.map((r) => (
            <button
              key={r.value}
              onClick={() => setRange(r.value)}
              className={[
                "px-4 py-1.5 rounded-lg text-sm font-medium transition-all",
                range === r.value ? "bg-white text-brand shadow-sm" : "text-gray-500 hover:text-gray-700",
              ].join(" ")}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-10 text-center text-gray-400 text-sm">Загрузка...</div>
        ) : orders.length === 0 ? (
          <div className="p-16 flex flex-col items-center text-gray-300 gap-2">
            <span className="text-4xl">🧾</span>
            <p className="text-sm">Нет заказов за выбранный период</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-[11px] text-gray-400 uppercase tracking-wide border-b border-gray-100 bg-gray-50/50">
                <th className="px-5 py-3 font-medium">Чек</th>
                <th className="px-5 py-3 font-medium">Время</th>
                <th className="px-5 py-3 font-medium">Статус</th>
                <th className="px-5 py-3 font-medium text-right">Сумма</th>
                {isAdmin && <th className="px-3 py-3 font-medium text-right">Действия</th>}
                <th className="px-2 py-3 w-8" />
              </tr>
            </thead>
            <tbody>
              {orders.map((order) => {
                const s = STATUS_LABELS[order.status] ?? { label: order.status, cls: "bg-gray-100 text-gray-500" };
                return (
                  <tr
                    key={order.id}
                    className="border-b border-gray-50 hover:bg-gray-50/60 cursor-pointer transition-colors"
                    onClick={() => setSelected(selected?.id === order.id ? null : order)}
                  >
                    <td className="px-5 py-3.5 font-mono font-semibold text-gray-800">
                      #{order.receipt_no}
                    </td>
                    <td className="px-5 py-3.5 text-gray-500">
                      {format(new Date(order.created_at), "HH:mm", { locale: ru })}
                    </td>
                    <td className="px-5 py-3.5">
                      <span className={`text-[11px] font-semibold px-2.5 py-0.5 rounded-full ${s.cls}`}>
                        {s.label}
                      </span>
                    </td>
                    <td className="px-5 py-3.5 text-right font-semibold text-gray-900">
                      {order.total.toLocaleString("ru-RU")} ₸
                    </td>
                    {isAdmin && (
                      <td className="px-3 py-3.5 text-right whitespace-nowrap">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            if (confirm(`Скрыть чек #${order.receipt_no} из списка? (данные сохранятся)`))
                              softDelete.mutate(order.id);
                          }}
                          className="text-[11px] font-semibold px-2 py-1 rounded-lg bg-gray-100 text-gray-600 hover:bg-gray-200 transition mr-1"
                          title="Скрыть (мягкое удаление)"
                        >
                          Скрыть
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            if (
                              confirm(
                                `Удалить чек #${order.receipt_no} НАВСЕГДА?\n\nЗаказ будет стёрт из БД, списанные ингредиенты вернутся на склад, бонусы клиента откатятся. Действие необратимо.`,
                              )
                            )
                              hardDelete.mutate(order.id);
                          }}
                          className="text-[11px] font-semibold px-2 py-1 rounded-lg bg-red-50 text-red-600 hover:bg-red-100 transition"
                          title="Удалить из БД (откатит склад и бонусы)"
                        >
                          Удалить
                        </button>
                      </td>
                    )}
                    <td className="px-2 py-3.5 text-gray-300">
                      {selected?.id === order.id ? "▲" : "▼"}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Expanded order detail */}
      {selected && (
        <div className="mt-4 bg-white rounded-2xl border border-gray-100 p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-gray-800">
              Чек #{selected.receipt_no}
            </h2>
            <button
              onClick={() => setSelected(null)}
              className="text-gray-400 hover:text-gray-600 text-xl leading-none"
            >
              ×
            </button>
          </div>

          {!detail ? (
            <p className="text-sm text-gray-400">Загрузка...</p>
          ) : (
            <div className="grid grid-cols-2 gap-6">
              <div>
                <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide mb-2">Позиции</p>
                <div className="space-y-1">
                  {(detail.items ?? []).map((item, i) => (
                    <div key={i} className="flex justify-between text-sm">
                      <span className="text-gray-700">
                        {item.product_name} × {item.qty}
                      </span>
                      <span className="font-medium text-gray-900">
                        {item.line_total.toLocaleString("ru-RU")} ₸
                      </span>
                    </div>
                  ))}
                </div>
                {detail.discount_total > 0 && (
                  <div className="flex justify-between text-sm mt-2 text-red-500">
                    <span>Скидка</span>
                    <span>−{detail.discount_total.toLocaleString("ru-RU")} ₸</span>
                  </div>
                )}
                <div className="flex justify-between text-sm font-bold mt-3 pt-3 border-t border-gray-100">
                  <span>Итого</span>
                  <span>{detail.total.toLocaleString("ru-RU")} ₸</span>
                </div>
              </div>

              <div>
                <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide mb-2">Оплата</p>
                <div className="space-y-1">
                  {(detail.payments ?? []).map((p, i) => (
                    <div key={i} className="flex justify-between text-sm">
                      <span className="text-gray-600">{p.method_name}</span>
                      <span className="font-medium">{p.amount.toLocaleString("ru-RU")} ₸</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
