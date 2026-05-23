import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { format, subDays, startOfDay } from "date-fns";
import { ru } from "date-fns/locale";
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend,
} from "recharts";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

interface DailyRevenueRow {
  day: string;
  orders_count: number;
  revenue: number;
  margin: number;
  avg_check: number;
  total_cost: number;
}

interface BaristaStatsRow {
  id: string;
  full_name: string;
  orders_count: number;
  revenue: number;
  avg_check: number;
  shifts_count: number;
}

interface ProductABCRow {
  product_id: string;
  product_name: string;
  category_name: string;
  units_sold: number;
  revenue: number;
  margin: number;
  margin_pct: number;
}

type Period = "week" | "month" | "quarter";
const PERIODS: { value: Period; label: string }[] = [
  { value: "week",    label: "Неделя" },
  { value: "month",   label: "Месяц" },
  { value: "quarter", label: "Квартал" },
];

function periodDates(p: Period) {
  const to = new Date();
  const days = p === "week" ? 6 : p === "month" ? 29 : 89;
  return { from: startOfDay(subDays(to, days)), to };
}

function tenge(n: number) { return n.toLocaleString("ru-RU") + " ₸"; }

export default function AnalyticsPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const qc = useQueryClient();
  const [period, setPeriod] = useState<Period>("month");
  const [refreshed, setRefreshed] = useState(false);

  const { from, to } = useMemo(() => periodDates(period), [period]);
  const fromISO = from.toISOString();
  const toISO   = to.toISOString();

  const { data: revenue = [] } = useQuery<DailyRevenueRow[]>({
    queryKey: ["revenue-detail", locationId, period],
    queryFn: () =>
      api.get("/analytics/revenue", { params: { location_id: locationId, from: fromISO, to: toISO } })
         .then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: baristas = [] } = useQuery<BaristaStatsRow[]>({
    queryKey: ["barista-stats", locationId, period],
    queryFn: () =>
      api.get("/analytics/baristas", { params: { location_id: locationId, from: fromISO, to: toISO } })
         .then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: abc = [] } = useQuery<ProductABCRow[]>({
    queryKey: ["abc-analytics", locationId],
    queryFn: () =>
      api.get("/analytics/products", { params: { location_id: locationId } })
         .then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const refreshMutation = useMutation({
    mutationFn: () => api.post("/analytics/refresh"),
    onSuccess: () => {
      setRefreshed(true);
      qc.invalidateQueries({ queryKey: ["revenue-detail"] });
      qc.invalidateQueries({ queryKey: ["barista-stats"] });
      qc.invalidateQueries({ queryKey: ["abc-analytics"] });
      setTimeout(() => setRefreshed(false), 2000);
    },
  });

  const chartData = revenue.map((r) => ({
    ...r,
    label: format(new Date(r.day), "dd.MM", { locale: ru }),
  }));

  const totalRevenue = revenue.reduce((s, r) => s + r.revenue, 0);
  const totalMargin  = revenue.reduce((s, r) => s + r.margin, 0);
  const totalOrders  = revenue.reduce((s, r) => s + r.orders_count, 0);
  const marginPct    = totalRevenue > 0 ? Math.round((totalMargin / totalRevenue) * 100) : 0;

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Аналитика</h1>
          <p className="text-sm text-gray-400 mt-0.5">
            {format(from, "d MMM", { locale: ru })} — {format(to, "d MMM yyyy", { locale: ru })}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => refreshMutation.mutate()}
            disabled={refreshMutation.isPending}
            className="text-sm text-gray-500 hover:text-brand transition-colors flex items-center gap-1.5 disabled:opacity-50"
          >
            <span className={refreshMutation.isPending ? "animate-spin inline-block" : ""}>🔄</span>
            {refreshed ? "Обновлено!" : "Обновить данные"}
          </button>
          <div className="flex gap-1 bg-gray-100 p-1 rounded-xl">
            {PERIODS.map((p) => (
              <button
                key={p.value}
                onClick={() => setPeriod(p.value)}
                className={[
                  "px-4 py-1.5 rounded-lg text-sm font-medium transition-all",
                  period === p.value ? "bg-white text-brand shadow-sm" : "text-gray-500 hover:text-gray-700",
                ].join(" ")}
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Summary KPIs */}
      <div className="grid grid-cols-4 gap-4">
        {[
          { label: "Выручка",     value: tenge(totalRevenue), accent: true },
          { label: "Маржа",       value: `${marginPct}%`,     sub: tenge(totalMargin) },
          { label: "Чеков",       value: totalOrders.toLocaleString("ru-RU") },
          { label: "Ср. чек",     value: tenge(totalOrders > 0 ? Math.round(totalRevenue / totalOrders) : 0) },
        ].map(({ label, value, sub, accent }) => (
          <div key={label} className="bg-white rounded-2xl border border-gray-100 p-5">
            <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-widest">{label}</p>
            <p className={`mt-2 text-2xl font-bold tabular-nums ${accent ? "text-brand" : "text-gray-900"}`}>
              {value}
            </p>
            {sub && <p className="mt-0.5 text-xs text-gray-400">{sub}</p>}
          </div>
        ))}
      </div>

      {/* Revenue + cost bar chart */}
      <div className="bg-white rounded-2xl border border-gray-100 p-6">
        <h2 className="text-sm font-semibold text-gray-800 mb-5">Выручка и себестоимость</h2>
        {chartData.length === 0 ? (
          <div className="h-52 flex items-center justify-center text-gray-300 text-sm">
            Нет данных за период
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={chartData} margin={{ top: 4, right: 4, left: 56, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#F3F4F6" vertical={false} />
              <XAxis dataKey="label" tick={{ fontSize: 11, fill: "#9CA3AF" }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fontSize: 11, fill: "#9CA3AF" }} axisLine={false} tickLine={false}
                tickFormatter={(v) => `${Math.round(v / 1000)}k`} />
              <Tooltip
                formatter={(v: number, name: string) => [
                  tenge(v),
                  name === "revenue" ? "Выручка" : name === "total_cost" ? "Себестоимость" : "Маржа",
                ]}
                contentStyle={{ borderRadius: 10, border: "1px solid #E5E7EB", fontSize: 12 }}
              />
              <Legend formatter={(v) =>
                v === "revenue" ? "Выручка" : v === "total_cost" ? "Себестоимость" : "Маржа"
              } wrapperStyle={{ fontSize: 12 }} />
              <Bar dataKey="revenue"    fill="#6B3F1F" radius={[4, 4, 0, 0]} />
              <Bar dataKey="total_cost" fill="#F0DCC9" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        )}
      </div>

      <div className="grid grid-cols-2 gap-6">
        {/* Barista stats */}
        <div className="bg-white rounded-2xl border border-gray-100 p-6">
          <h2 className="text-sm font-semibold text-gray-800 mb-4">По бариста</h2>
          {baristas.length === 0 ? (
            <div className="h-40 flex items-center justify-center text-gray-300 text-sm">Нет данных</div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[11px] text-gray-400 uppercase tracking-wide border-b border-gray-100">
                  <th className="pb-2 text-left font-medium">Бариста</th>
                  <th className="pb-2 text-right font-medium">Чеков</th>
                  <th className="pb-2 text-right font-medium">Выручка</th>
                  <th className="pb-2 text-right font-medium">Ср. чек</th>
                </tr>
              </thead>
              <tbody>
                {baristas.map((b) => (
                  <tr key={b.id} className="border-b border-gray-50">
                    <td className="py-3 font-medium text-gray-900">{b.full_name}</td>
                    <td className="py-3 text-right text-gray-600 tabular-nums">{b.orders_count}</td>
                    <td className="py-3 text-right font-semibold tabular-nums">{tenge(b.revenue)}</td>
                    <td className="py-3 text-right text-gray-600 tabular-nums">{tenge(Math.round(b.avg_check))}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* ABC top 10 */}
        <div className="bg-white rounded-2xl border border-gray-100 p-6">
          <h2 className="text-sm font-semibold text-gray-800 mb-4">Топ-10 ABC</h2>
          {abc.length === 0 ? (
            <div className="h-40 flex items-center justify-center text-gray-300 text-sm">Нет данных</div>
          ) : (
            <div className="space-y-2">
              {abc.slice(0, 10).map((row, i) => {
                const share = abc[0].revenue > 0 ? (row.revenue / abc[0].revenue) * 100 : 0;
                const abcClass =
                  i < 3 ? "bg-emerald-500" : i < 6 ? "bg-amber-400" : "bg-gray-200";
                return (
                  <div key={row.product_id} className="flex items-center gap-3">
                    <span className="w-4 text-xs text-gray-300 text-right font-mono">{i + 1}</span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between mb-0.5">
                        <span className="text-sm font-medium text-gray-800 truncate">{row.product_name}</span>
                        <span className="text-xs font-semibold text-gray-600 ml-2 shrink-0">
                          {tenge(row.revenue)}
                        </span>
                      </div>
                      <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full ${abcClass}`}
                          style={{ width: `${share}%` }}
                        />
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
