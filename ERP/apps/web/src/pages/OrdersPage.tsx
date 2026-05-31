import { Fragment, useMemo, useState } from "react";
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
  { value: "week", label: "Неделя" },
  { value: "month", label: "Месяц" },
];

const STATUS_LABELS: Record<string, { label: string; cls: string }> = {
  paid: { label: "Оплачен", cls: "bg-emerald-50 text-emerald-700" },
  pending: { label: "Ожидание", cls: "bg-amber-50 text-amber-700" },
  cancelled: { label: "Отменён", cls: "bg-red-50 text-red-600" },
};

function rangeDates(r: Range) {
  const to = new Date();
  const from =
    r === "today"
      ? startOfDay(to)
      : r === "week"
        ? startOfDay(subDays(to, 6))
        : startOfDay(subDays(to, 29));
  return { from, to };
}

// Bucket key — YYYY-MM-DD in local time so we group by calendar day
// regardless of UTC offset.
function dayKey(iso: string): string {
  const d = new Date(iso);
  return format(d, "yyyy-MM-dd");
}

export default function OrdersPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const isAdmin = user?.role === "admin";
  const qc = useQueryClient();
  const [range, setRange] = useState<Range>("today");
  // Which order's details are expanded inline under its row.
  const [expandedOrderId, setExpandedOrderId] = useState<string | null>(null);
  // Which day buckets are open (week/month view). Empty = all collapsed.
  const [openDays, setOpenDays] = useState<Record<string, boolean>>({});

  const refreshOrders = () => {
    qc.invalidateQueries({ queryKey: ["orders", locationId, range] });
    setExpandedOrderId(null);
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
    queryKey: ["order", expandedOrderId],
    queryFn: () => api.get(`/orders/${expandedOrderId!}`).then((r) => r.data.data),
    enabled: !!expandedOrderId,
  });

  // Group orders by calendar day for week / month views. Stable order:
  // most-recent day first, orders within a day already arrive newest-first
  // from the backend.
  const daysGrouped = useMemo(() => {
    const map = new Map<string, OrderDTO[]>();
    for (const o of orders) {
      const k = dayKey(o.created_at);
      const bucket = map.get(k) ?? [];
      bucket.push(o);
      map.set(k, bucket);
    }
    return Array.from(map.entries()).sort((a, b) => (a[0] < b[0] ? 1 : -1));
  }, [orders]);

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
              onClick={() => {
                setRange(r.value);
                setExpandedOrderId(null);
                setOpenDays({});
              }}
              className={[
                "px-4 py-1.5 rounded-lg text-sm font-medium transition-all",
                range === r.value
                  ? "bg-white text-brand shadow-sm"
                  : "text-gray-500 hover:text-gray-700",
              ].join(" ")}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <div className="bg-white rounded-2xl border border-gray-100 p-10 text-center text-gray-400 text-sm">
          Загрузка...
        </div>
      ) : orders.length === 0 ? (
        <div className="bg-white rounded-2xl border border-gray-100 p-16 flex flex-col items-center text-gray-300 gap-2">
          <span className="text-4xl">🧾</span>
          <p className="text-sm">Нет заказов за выбранный период</p>
        </div>
      ) : range === "today" ? (
        // Flat list — chronological, single day.
        <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
          <OrdersTable
            orders={orders}
            expandedOrderId={expandedOrderId}
            setExpandedOrderId={setExpandedOrderId}
            detail={detail}
            isAdmin={isAdmin}
            onSoftDelete={(id) => softDelete.mutate(id)}
            onHardDelete={(id) => hardDelete.mutate(id)}
          />
        </div>
      ) : (
        // Grouped by day. Each bucket collapses; if every bucket is open we
        // effectively show one continuous list.
        <div className="space-y-3">
          {daysGrouped.map(([day, dayOrders]) => {
            const isOpen = !!openDays[day];
            const dayRevenue = dayOrders.reduce(
              (s, o) => s + (o.status === "paid" ? o.total : 0),
              0
            );
            const dayPaid = dayOrders.filter((o) => o.status === "paid").length;
            return (
              <div
                key={day}
                className="bg-white rounded-2xl border border-gray-100 overflow-hidden"
              >
                <button
                  onClick={() => setOpenDays((prev) => ({ ...prev, [day]: !prev[day] }))}
                  className="w-full px-5 py-3.5 flex items-center justify-between hover:bg-gray-50/60 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <span className="text-gray-300 text-sm">{isOpen ? "▼" : "▶"}</span>
                    <div className="text-left">
                      <p className="font-semibold text-gray-900">
                        {format(new Date(day), "EEEE, d MMMM", { locale: ru })}
                      </p>
                      <p className="text-xs text-gray-400 mt-0.5">{dayPaid} чеков</p>
                    </div>
                  </div>
                  <span className="font-semibold text-gray-900 tabular-nums">
                    {dayRevenue.toLocaleString("ru-RU")} ₸
                  </span>
                </button>
                {isOpen && (
                  <OrdersTable
                    orders={dayOrders}
                    expandedOrderId={expandedOrderId}
                    setExpandedOrderId={setExpandedOrderId}
                    detail={detail}
                    isAdmin={isAdmin}
                    onSoftDelete={(id) => softDelete.mutate(id)}
                    onHardDelete={(id) => hardDelete.mutate(id)}
                  />
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

// Renders a list of orders with inline expansion under each clicked row.
// Used both standalone (today view) and inside each day group (week/month).
function OrdersTable({
  orders,
  expandedOrderId,
  setExpandedOrderId,
  detail,
  isAdmin,
  onSoftDelete,
  onHardDelete,
}: {
  orders: OrderDTO[];
  expandedOrderId: string | null;
  setExpandedOrderId: (id: string | null) => void;
  detail: OrderDTO | undefined;
  isAdmin: boolean;
  onSoftDelete: (id: string) => void;
  onHardDelete: (id: string) => void;
}) {
  return (
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
          const s = STATUS_LABELS[order.status] ?? {
            label: order.status,
            cls: "bg-gray-100 text-gray-500",
          };
          const isExpanded = expandedOrderId === order.id;
          return (
            <Fragment key={order.id}>
              <tr
                className="border-b border-gray-50 hover:bg-gray-50/60 cursor-pointer transition-colors"
                onClick={() => setExpandedOrderId(isExpanded ? null : order.id)}
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
                        if (
                          confirm(`Скрыть чек #${order.receipt_no} из списка? (данные сохранятся)`)
                        )
                          onSoftDelete(order.id);
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
                            `Удалить чек #${order.receipt_no} НАВСЕГДА?\n\nЗаказ будет стёрт из БД, списанные ингредиенты вернутся на склад, бонусы клиента откатятся. Действие необратимо.`
                          )
                        )
                          onHardDelete(order.id);
                      }}
                      className="text-[11px] font-semibold px-2 py-1 rounded-lg bg-red-50 text-red-600 hover:bg-red-100 transition"
                      title="Удалить из БД (откатит склад и бонусы)"
                    >
                      Удалить
                    </button>
                  </td>
                )}
                <td className="px-2 py-3.5 text-gray-300">{isExpanded ? "▲" : "▼"}</td>
              </tr>
              {isExpanded && (
                <tr className="bg-gray-50/40">
                  <td colSpan={isAdmin ? 6 : 5} className="px-5 py-4 border-b border-gray-100">
                    <OrderDetailCard
                      summary={order}
                      detail={detail?.id === order.id ? detail : undefined}
                    />
                  </td>
                </tr>
              )}
            </Fragment>
          );
        })}
      </tbody>
    </table>
  );
}

// Receipt panel rendered directly under the clicked row. Falls back to a
// "loading" line while the per-order detail query is in flight.
function OrderDetailCard({ summary, detail }: { summary: OrderDTO; detail: OrderDTO | undefined }) {
  if (!detail) {
    return <p className="text-sm text-gray-400">Загрузка чека #{summary.receipt_no}...</p>;
  }
  return (
    <div className="grid grid-cols-2 gap-6">
      <div>
        <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide mb-2">
          Позиции
        </p>
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
        <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide mb-2">
          Оплата
        </p>
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
  );
}
