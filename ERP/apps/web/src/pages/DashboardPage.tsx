import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { format, subDays, startOfDay } from "date-fns";
import { ru } from "date-fns/locale";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

// ── Types matching analytics_service.go DTOs ──────────────────────────────────

interface DailyRevenueRow {
  day: string;
  orders_count: number;
  revenue: number;
  margin: number;
  avg_check: number;
  total_cost: number;
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

interface HeatmapRow {
  weekday: number;
  hour_of_day: number;
  orders_count: number;
  revenue: number;
}

// ── Period logic ──────────────────────────────────────────────────────────────

type Period = "today" | "week" | "month";

const PERIODS: { value: Period; label: string }[] = [
  { value: "today", label: "Сегодня" },
  { value: "week",  label: "Неделя" },
  { value: "month", label: "Месяц" },
];

function periodDates(p: Period): { from: Date; to: Date } {
  const to = new Date();
  const from =
    p === "today" ? startOfDay(to) :
    p === "week"  ? startOfDay(subDays(to, 6)) :
                    startOfDay(subDays(to, 29));
  return { from, to };
}

// ── Helpers ───────────────────────────────────────────────────────────────────

const DAYS_SHORT = ["Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"];
const HOURS = Array.from({ length: 24 }, (_, i) => i);

function tenge(n: number): string {
  return n.toLocaleString("ru-RU") + " ₸";
}

function heatColor(count: number, max: number): string {
  if (!max || !count) return "#F3F4F6";
  const t = count / max;
  return `rgba(107, 63, 31, ${0.12 + t * 0.78})`;
}

// ── Dashboard ─────────────────────────────────────────────────────────────────

export default function DashboardPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const [period, setPeriod] = useState<Period>("week");

  const { from, to } = useMemo(() => periodDates(period), [period]);
  const fromISO = from.toISOString();
  const toISO   = to.toISOString();

  const { data: revenue = [], isLoading: loadingKpi } = useQuery<DailyRevenueRow[]>({
    queryKey: ["revenue", locationId, period],
    queryFn: () =>
      api
        .get("/analytics/revenue", { params: { location_id: locationId, from: fromISO, to: toISO } })
        .then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: abc = [] } = useQuery<ProductABCRow[]>({
    queryKey: ["abc", locationId],
    queryFn: () =>
      api
        .get("/analytics/products", { params: { location_id: locationId } })
        .then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: heatmap = [] } = useQuery<HeatmapRow[]>({
    queryKey: ["heatmap", locationId, period],
    queryFn: () =>
      api
        .get("/analytics/heatmap", { params: { location_id: locationId, from: fromISO, to: toISO } })
        .then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  // KPIs
  const totalRevenue = revenue.reduce((s, r) => s + r.revenue, 0);
  const totalOrders  = revenue.reduce((s, r) => s + r.orders_count, 0);
  const avgCheck     = totalOrders > 0 ? Math.round(totalRevenue / totalOrders) : 0;
  const totalMargin  = revenue.reduce((s, r) => s + r.margin, 0);
  const marginPct    = totalRevenue > 0 ? Math.round((totalMargin / totalRevenue) * 100) : 0;

  // Chart data
  const chartData = revenue.map((r) => ({
    ...r,
    label: format(new Date(r.day), period === "today" ? "HH:mm" : "dd.MM", { locale: ru }),
  }));

  // Heatmap lookup
  const heatMap = useMemo(() => {
    const m: Record<string, number> = {};
    heatmap.forEach((r) => { m[`${r.weekday}-${r.hour_of_day}`] = r.orders_count; });
    return m;
  }, [heatmap]);
  const maxCount = Math.max(0, ...heatmap.map((r) => r.orders_count));

  return (
    <div className="p-6 space-y-6 max-w-[1400px] mx-auto">
      {/* ── Header ── */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Дашборд</h1>
          <p className="text-sm text-gray-400 mt-0.5">
            {format(from, "d MMM", { locale: ru })} — {format(to, "d MMM yyyy", { locale: ru })}
          </p>
        </div>

        <div className="flex gap-1 bg-gray-100 p-1 rounded-xl">
          {PERIODS.map((p) => (
            <button
              key={p.value}
              onClick={() => setPeriod(p.value)}
              className={[
                "px-4 py-1.5 rounded-lg text-sm font-medium transition-all",
                period === p.value
                  ? "bg-white text-brand shadow-sm"
                  : "text-gray-500 hover:text-gray-700",
              ].join(" ")}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {/* ── KPI Cards ── */}
      <div className="grid grid-cols-4 gap-4">
        <KpiCard label="Выручка"     value={tenge(totalRevenue)}               accent loading={loadingKpi} />
        <KpiCard label="Чеков"       value={totalOrders.toLocaleString("ru-RU")}       loading={loadingKpi} />
        <KpiCard label="Средний чек" value={tenge(avgCheck)}                           loading={loadingKpi} />
        <KpiCard label="Маржа"       value={`${marginPct}%`} sub={tenge(totalMargin)} loading={loadingKpi} />
      </div>

      {/* ── Revenue chart ── */}
      <div className="bg-white rounded-2xl border border-gray-100 p-6">
        <h2 className="text-sm font-semibold text-gray-800 mb-5">Выручка по дням</h2>
        {chartData.length === 0 ? (
          <Blank h={220} />
        ) : (
          <ResponsiveContainer width="100%" height={220}>
            <AreaChart data={chartData} margin={{ top: 4, right: 4, left: 56, bottom: 0 }}>
              <defs>
                <linearGradient id="grad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%"  stopColor="#6B3F1F" stopOpacity={0.15} />
                  <stop offset="95%" stopColor="#6B3F1F" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#F3F4F6" vertical={false} />
              <XAxis
                dataKey="label"
                tick={{ fontSize: 11, fill: "#9CA3AF" }}
                axisLine={false} tickLine={false}
              />
              <YAxis
                tick={{ fontSize: 11, fill: "#9CA3AF" }}
                axisLine={false} tickLine={false}
                tickFormatter={(v) => `${Math.round(v / 1000)}k`}
              />
              <Tooltip
                formatter={(v: number) => [tenge(v), "Выручка"]}
                contentStyle={{ borderRadius: 10, border: "1px solid #E5E7EB", fontSize: 12 }}
                labelStyle={{ fontWeight: 600, color: "#374151" }}
              />
              <Area
                type="monotone"
                dataKey="revenue"
                stroke="#6B3F1F"
                strokeWidth={2}
                fill="url(#grad)"
                dot={false}
                activeDot={{ r: 4, fill: "#6B3F1F" }}
              />
            </AreaChart>
          </ResponsiveContainer>
        )}
      </div>

      {/* ── Bottom row: Heatmap + ABC ── */}
      <div className="grid grid-cols-5 gap-6">
        {/* Heatmap */}
        <div className="col-span-2 bg-white rounded-2xl border border-gray-100 p-6">
          <h2 className="text-sm font-semibold text-gray-800 mb-4">Тепловая карта</h2>

          {/* Day labels */}
          <div className="flex mb-1 pl-7">
            {DAYS_SHORT.map((d) => (
              <div key={d} className="w-7 text-center text-[10px] text-gray-400 font-medium">
                {d}
              </div>
            ))}
          </div>

          {/* Hour rows */}
          {HOURS.map((h) => (
            <div key={h} className="flex items-center mb-[2px]">
              <span className="w-6 text-right text-[10px] text-gray-300 mr-1 shrink-0">
                {h % 3 === 0 ? `${h}` : ""}
              </span>
              {[0, 1, 2, 3, 4, 5, 6].map((d) => {
                const cnt = heatMap[`${d}-${h}`] ?? 0;
                return (
                  <div
                    key={d}
                    className="w-7 h-[18px] mx-[1px] rounded-[3px] transition-colors"
                    style={{ backgroundColor: heatColor(cnt, maxCount) }}
                    title={`${DAYS_SHORT[d]} ${h}:00 — ${cnt} заказ.`}
                  />
                );
              })}
            </div>
          ))}

          <div className="flex items-center gap-2 mt-3 ml-7">
            <span className="text-[10px] text-gray-400">0</span>
            <div className="h-2 flex-1 rounded-full" style={{
              background: "linear-gradient(to right, #F3F4F6, rgba(107,63,31,0.9))"
            }} />
            <span className="text-[10px] text-gray-400">{maxCount}</span>
          </div>
        </div>

        {/* ABC Table */}
        <div className="col-span-3 bg-white rounded-2xl border border-gray-100 p-6">
          <h2 className="text-sm font-semibold text-gray-800 mb-4">ABC-анализ продуктов</h2>

          {abc.length === 0 ? (
            <Blank h={200} label="Нет данных о продажах" />
          ) : (
            <div className="overflow-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-[11px] text-gray-400 uppercase tracking-wide border-b border-gray-100">
                    <th className="pb-2.5 font-medium w-7">#</th>
                    <th className="pb-2.5 font-medium">Товар</th>
                    <th className="pb-2.5 font-medium text-right">Продаж</th>
                    <th className="pb-2.5 font-medium text-right">Выручка</th>
                    <th className="pb-2.5 font-medium text-right">Маржа</th>
                  </tr>
                </thead>
                <tbody>
                  {abc.slice(0, 12).map((row, i) => (
                    <tr
                      key={row.product_id}
                      className="border-b border-gray-50 hover:bg-gray-50/60 transition-colors"
                    >
                      <td className="py-3 text-gray-300 text-xs font-mono">{i + 1}</td>
                      <td className="py-3">
                        <p className="font-medium text-gray-900 leading-tight">{row.product_name}</p>
                        <p className="text-[11px] text-gray-400 leading-tight mt-0.5">
                          {row.category_name}
                        </p>
                      </td>
                      <td className="py-3 text-right text-gray-600 tabular-nums">
                        {row.units_sold}
                      </td>
                      <td className="py-3 text-right font-semibold text-gray-900 tabular-nums">
                        {tenge(row.revenue)}
                      </td>
                      <td className="py-3 text-right">
                        <MarginBadge pct={row.margin_pct} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ── Sub-components ────────────────────────────────────────────────────────────

function KpiCard({
  label, value, sub, loading, accent,
}: {
  label: string;
  value: string;
  sub?: string;
  loading?: boolean;
  accent?: boolean;
}) {
  return (
    <div className="bg-white rounded-2xl border border-gray-100 p-5">
      <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-widest">{label}</p>
      {loading ? (
        <div className="mt-3 h-8 w-28 bg-gray-100 rounded-lg animate-pulse" />
      ) : (
        <p className={["mt-2 text-2xl font-bold tabular-nums", accent ? "text-brand" : "text-gray-900"].join(" ")}>
          {value}
        </p>
      )}
      {sub && <p className="mt-0.5 text-xs text-gray-400">{sub}</p>}
    </div>
  );
}

function MarginBadge({ pct }: { pct: number }) {
  const cls =
    pct >= 50 ? "bg-emerald-50 text-emerald-700" :
    pct >= 25 ? "bg-amber-50 text-amber-700" :
                "bg-red-50 text-red-600";
  return (
    <span className={["inline-block text-[11px] font-semibold px-2 py-0.5 rounded-full tabular-nums", cls].join(" ")}>
      {pct.toFixed(0)}%
    </span>
  );
}

function Blank({ h, label = "Нет данных за выбранный период" }: { h: number; label?: string }) {
  return (
    <div
      className="flex flex-col items-center justify-center text-gray-300 gap-2"
      style={{ height: h }}
    >
      <span className="text-3xl">📭</span>
      <p className="text-sm">{label}</p>
    </div>
  );
}
