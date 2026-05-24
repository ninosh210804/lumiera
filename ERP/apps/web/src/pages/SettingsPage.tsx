import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

interface UserListItem {
  id: string;
  full_name: string;
  email: string | null;
  role: string;
  is_active: boolean;
  location_id: string;
}

interface RoleRow {
  id: string;
  code: string;
  name: string;
}

interface NewUserForm {
  full_name: string;
  role_id: string;
  email: string;
  pin: string;
  pin_confirm: string;
}

const ROLE_LABELS: Record<string, { label: string; cls: string }> = {
  admin:   { label: "Администратор", cls: "bg-brand-50 text-brand font-semibold" },
  manager: { label: "Менеджер",      cls: "bg-blue-50 text-blue-700 font-semibold" },
  barista: { label: "Бариста",       cls: "bg-gray-100 text-gray-600" },
};

export default function SettingsPage() {
  const { user } = useAuth();
  const locationId = user?.location_id ?? "";
  const qc = useQueryClient();

  const [showAddUser, setShowAddUser] = useState(false);
  const [changePinFor, setChangePinFor] = useState<UserListItem | null>(null);
  const [newPin, setNewPin] = useState("");
  const [pinError, setPinError] = useState("");

  const { data: users = [], isLoading } = useQuery<UserListItem[]>({
    queryKey: ["users", locationId],
    queryFn: () =>
      api.get("/users", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const { data: roles = [] } = useQuery<RoleRow[]>({
    queryKey: ["roles"],
    queryFn: () => api.get("/roles").then((r) => r.data.data ?? []),
    enabled: !!locationId,
  });

  const [form, setForm] = useState<NewUserForm>({
    full_name: "", role_id: "", email: "", pin: "", pin_confirm: "",
  });
  const [formError, setFormError] = useState("");

  const createMutation = useMutation({
    mutationFn: (payload: object) => api.post("/users", payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["users", locationId] });
      setShowAddUser(false);
      setForm({ full_name: "", role_id: "", email: "", pin: "", pin_confirm: "" });
      setFormError("");
    },
    onError: (err: unknown) => {
      const msg = (err as { response?: { data?: { error?: string } } }).response?.data?.error;
      setFormError(msg ?? "Ошибка при создании сотрудника");
    },
  });

  const deactivateMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/users/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["users", locationId] }),
  });

  const pinMutation = useMutation({
    mutationFn: ({ id, pin }: { id: string; pin: string }) =>
      api.put(`/users/${id}/pin`, { pin }),
    onSuccess: () => {
      setChangePinFor(null);
      setNewPin("");
      setPinError("");
    },
    onError: () => setPinError("Ошибка при смене PIN"),
  });

  const activeUsers   = users.filter((u) => u.is_active);
  const inactiveUsers = users.filter((u) => !u.is_active);

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Настройки</h1>
        <button
          onClick={() => setShowAddUser(true)}
          className="text-sm bg-brand text-white px-4 py-2 rounded-xl font-semibold hover:bg-brand-600 transition-colors"
        >
          + Добавить сотрудника
        </button>
      </div>

      {/* Users */}
      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-100 bg-gray-50/40">
          <h2 className="text-sm font-semibold text-gray-700">
            Сотрудники ({activeUsers.length} активных)
          </h2>
        </div>

        {isLoading ? (
          <div className="p-10 text-center text-sm text-gray-400">Загрузка...</div>
        ) : users.length === 0 ? (
          <div className="p-16 flex flex-col items-center text-gray-300 gap-2">
            <span className="text-4xl">👥</span>
            <p className="text-sm">Нет сотрудников</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-[11px] text-gray-400 uppercase tracking-wide border-b border-gray-100">
                <th className="px-6 py-3 text-left font-medium">Имя</th>
                <th className="px-6 py-3 text-left font-medium">Email</th>
                <th className="px-6 py-3 text-left font-medium">Роль</th>
                <th className="px-6 py-3 text-left font-medium">Статус</th>
                <th className="px-6 py-3 text-right font-medium">Действия</th>
              </tr>
            </thead>
            <tbody>
              {[...activeUsers, ...inactiveUsers].map((u) => {
                const role = ROLE_LABELS[u.role] ?? { label: u.role, cls: "bg-gray-100 text-gray-500" };
                const isMe = u.id === user?.id;
                return (
                  <tr key={u.id} className={["border-b border-gray-50 transition-colors",
                    u.is_active ? "hover:bg-gray-50/60" : "opacity-50 bg-gray-50/30"].join(" ")}>
                    <td className="px-6 py-3.5">
                      <div className="flex items-center gap-2.5">
                        <div className="w-7 h-7 rounded-full bg-brand-100 text-brand text-xs font-bold flex items-center justify-center shrink-0">
                          {u.full_name[0]?.toUpperCase()}
                        </div>
                        <span className="font-medium text-gray-900">
                          {u.full_name}
                          {isMe && <span className="text-gray-400 font-normal"> (я)</span>}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-3.5 text-gray-500">{u.email ?? "—"}</td>
                    <td className="px-6 py-3.5">
                      <span className={`text-[11px] px-2.5 py-0.5 rounded-full ${role.cls}`}>
                        {role.label}
                      </span>
                    </td>
                    <td className="px-6 py-3.5">
                      <span className={u.is_active
                        ? "text-emerald-600 text-xs font-medium"
                        : "text-gray-400 text-xs"}>
                        {u.is_active ? "● Активен" : "○ Деактивирован"}
                      </span>
                    </td>
                    <td className="px-6 py-3.5 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => { setChangePinFor(u); setNewPin(""); setPinError(""); }}
                          className="text-xs text-blue-500 hover:text-blue-700 font-medium transition-colors"
                        >
                          Сменить PIN
                        </button>
                        {u.is_active && !isMe && user?.role === "admin" && (
                          <button
                            onClick={() => {
                              if (confirm(`Деактивировать ${u.full_name}?`)) {
                                deactivateMutation.mutate(u.id);
                              }
                            }}
                            className="text-xs text-red-400 hover:text-red-600 font-medium transition-colors"
                          >
                            Деактивировать
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Loyalty / promo */}
      <LoyaltyPromoCard canEdit={user?.role === "admin" || user?.role === "manager"} />

      {/* System info card */}
      <div className="bg-white rounded-2xl border border-gray-100 p-6">
        <h2 className="text-sm font-semibold text-gray-700 mb-4">Система</h2>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <InfoRow label="Локация ID" value={locationId.slice(0, 8) + "..."} />
          <InfoRow label="Роль"       value={user?.role ?? "—"} />
          <InfoRow label="API"        value={window.location.origin + "/api/v1"} />
          <InfoRow label="Версия"     value="1.0.0" />
        </div>
      </div>

      {/* Change PIN modal */}
      {changePinFor && (
        <Modal title={`Смена PIN — ${changePinFor.full_name}`} onClose={() => setChangePinFor(null)}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Новый PIN</label>
              <input
                type="password"
                value={newPin}
                onChange={(e) => setNewPin(e.target.value)}
                maxLength={6}
                placeholder="4–6 цифр"
                className="w-full px-3 py-2 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
                autoFocus
              />
              {pinError && <p className="text-xs text-red-500 mt-1">{pinError}</p>}
            </div>
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setChangePinFor(null)}
                className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700 transition"
              >
                Отмена
              </button>
              <button
                disabled={newPin.length < 4 || pinMutation.isPending}
                onClick={() => pinMutation.mutate({ id: changePinFor.id, pin: newPin })}
                className="px-4 py-2 text-sm bg-brand text-white rounded-xl font-semibold hover:bg-brand-600 transition disabled:opacity-50"
              >
                Сохранить
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* Add user form */}
      {showAddUser && (
        <Modal
          title="Новый сотрудник"
          onClose={() => {
            setShowAddUser(false);
            setForm({ full_name: "", role_id: "", email: "", pin: "", pin_confirm: "" });
            setFormError("");
          }}
        >
          <div className="space-y-4">
            {/* Full name */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Имя и фамилия <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={form.full_name}
                onChange={(e) => setForm((f) => ({ ...f, full_name: e.target.value }))}
                placeholder="Айдана Бекова"
                className="w-full px-3 py-2 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
                autoFocus
              />
            </div>

            {/* Role */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Роль <span className="text-red-400">*</span>
              </label>
              <select
                value={form.role_id}
                onChange={(e) => setForm((f) => ({ ...f, role_id: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/30 bg-white"
              >
                <option value="">Выберите роль...</option>
                {roles.map((r) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
            </div>

            {/* Email (optional for all roles) */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Email
                <span className="text-gray-400 font-normal ml-1">(для входа по email+PIN)</span>
              </label>
              <input
                type="email"
                value={form.email}
                onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
                placeholder="email@coffeeshop.kz"
                className="w-full px-3 py-2 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
              />
            </div>

            {/* PIN */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  PIN <span className="text-red-400">*</span>
                </label>
                <input
                  type="password"
                  value={form.pin}
                  onChange={(e) => setForm((f) => ({ ...f, pin: e.target.value }))}
                  placeholder="4–6 цифр"
                  maxLength={6}
                  className="w-full px-3 py-2 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/30"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Повтор PIN <span className="text-red-400">*</span>
                </label>
                <input
                  type="password"
                  value={form.pin_confirm}
                  onChange={(e) => setForm((f) => ({ ...f, pin_confirm: e.target.value }))}
                  placeholder="••••"
                  maxLength={6}
                  className={[
                    "w-full px-3 py-2 border rounded-xl text-sm focus:outline-none focus:ring-2",
                    form.pin_confirm && form.pin !== form.pin_confirm
                      ? "border-red-300 focus:ring-red-200"
                      : "border-gray-200 focus:ring-brand/30",
                  ].join(" ")}
                />
              </div>
            </div>
            {form.pin_confirm && form.pin !== form.pin_confirm && (
              <p className="text-xs text-red-500 -mt-2">PIN-коды не совпадают</p>
            )}

            {formError && (
              <p className="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-xl">{formError}</p>
            )}

            <div className="flex gap-2 justify-end pt-1">
              <button
                onClick={() => {
                  setShowAddUser(false);
                  setForm({ full_name: "", role_id: "", email: "", pin: "", pin_confirm: "" });
                  setFormError("");
                }}
                className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700 transition"
              >
                Отмена
              </button>
              <button
                disabled={
                  !form.full_name ||
                  !form.role_id ||
                  form.pin.length < 4 ||
                  form.pin !== form.pin_confirm ||
                  createMutation.isPending
                }
                onClick={() => {
                  setFormError("");
                  createMutation.mutate({
                    location_id: locationId,
                    role_id: form.role_id,
                    full_name: form.full_name,
                    email: form.email || undefined,
                    pin: form.pin,
                  });
                }}
                className="px-5 py-2 text-sm bg-brand text-white rounded-xl font-semibold hover:bg-brand-600 transition disabled:opacity-40"
              >
                {createMutation.isPending ? "Создание..." : "Создать"}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}

interface LoyaltyConfig {
  promo_active: boolean;
  promo_percent: number;
  every_n: number;
  free_category: string;
}

function LoyaltyPromoCard({ canEdit }: { canEdit: boolean }) {
  const qc = useQueryClient();
  const { data: cfg } = useQuery<LoyaltyConfig>({
    queryKey: ["loyalty-config"],
    queryFn: () => api.get("/loyalty/config").then((r) => r.data.data),
  });
  const [percent, setPercent] = useState<string>("");

  const setPromo = useMutation({
    mutationFn: (body: { active: boolean; percent: number }) => api.post("/loyalty/promo", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["loyalty-config"] }),
  });

  const pct = percent !== "" ? Number(percent) : cfg?.promo_percent ?? 10;

  return (
    <div className="bg-white rounded-2xl border border-gray-100 p-6">
      <h2 className="text-sm font-semibold text-gray-700 mb-4">Акции и бонусы</h2>
      <div className="space-y-4 text-sm">
        <div className="flex items-center justify-between">
          <div>
            <p className="font-medium text-gray-800">Скидка на всё</p>
            <p className="text-xs text-gray-400">
              Применяется ко всем заказам, пока акция включена
            </p>
          </div>
          <div className="flex items-center gap-2">
            <input
              type="number"
              min="0"
              max="100"
              value={percent !== "" ? percent : cfg?.promo_percent ?? ""}
              onChange={(e) => setPercent(e.target.value)}
              disabled={!canEdit}
              className="w-16 px-2 py-1.5 border border-gray-200 rounded-lg text-sm text-right disabled:bg-gray-50"
            />
            <span className="text-gray-400">%</span>
            <button
              onClick={() => setPromo.mutate({ active: !cfg?.promo_active, percent: pct })}
              disabled={!canEdit || setPromo.isPending}
              className={[
                "px-3 py-1.5 rounded-lg text-xs font-semibold transition disabled:opacity-50",
                cfg?.promo_active
                  ? "bg-emerald-100 text-emerald-700 hover:bg-emerald-200"
                  : "bg-gray-100 text-gray-500 hover:bg-gray-200",
              ].join(" ")}
            >
              {cfg?.promo_active ? "ВКЛ" : "ВЫКЛ"}
            </button>
          </div>
        </div>

        {canEdit && percent !== "" && Number(percent) !== (cfg?.promo_percent ?? 10) && (
          <button
            onClick={() => setPromo.mutate({ active: cfg?.promo_active ?? false, percent: pct })}
            className="text-xs font-semibold text-brand hover:underline"
          >
            Сохранить {pct}%
          </button>
        )}

        <div className="border-t border-gray-100 pt-3 flex items-center justify-between">
          <div>
            <p className="font-medium text-gray-800">Каждый N-й кофе бесплатно</p>
            <p className="text-xs text-gray-400">
              По номеру телефона клиента · категория «{cfg?.free_category ?? "Кофе"}»
            </p>
          </div>
          <span className="bg-amber-50 text-amber-700 px-3 py-1.5 rounded-lg text-xs font-semibold">
            каждый {cfg?.every_n ?? 7}-й
          </span>
        </div>
      </div>
    </div>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-[11px] font-semibold text-gray-400 uppercase tracking-wide">{label}</p>
      <p className="mt-0.5 text-gray-700 font-mono text-xs break-all">{value}</p>
    </div>
  );
}

function Modal({ title, children, onClose }: { title: string; children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm">
      <div className="bg-white rounded-2xl shadow-xl p-6 w-full max-w-sm mx-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-semibold text-gray-900">{title}</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl leading-none">×</button>
        </div>
        {children}
      </div>
    </div>
  );
}
