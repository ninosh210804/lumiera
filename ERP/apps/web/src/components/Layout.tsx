import { NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../lib/auth";

interface NavItem {
  to: string;
  icon: string;
  label: string;
  allowedRoles?: string[];
}

const NAV: NavItem[] = [
  { to: "/dashboard", icon: "📊", label: "Дашборд", allowedRoles: ["admin", "manager"] },
  { to: "/pos", icon: "🛒", label: "Касса (POS)", allowedRoles: ["barista", "manager", "admin"] },
  { to: "/orders", icon: "🧾", label: "Заказы", allowedRoles: ["admin", "manager", "barista"] },
  { to: "/menu", icon: "☕", label: "Меню", allowedRoles: ["admin", "manager"] },
  { to: "/sales", icon: "🏷️", label: "Акции", allowedRoles: ["admin", "manager"] },
  { to: "/inventory", icon: "📦", label: "Склад (просмотр)", allowedRoles: ["admin", "manager"] },
  { to: "/warehouse", icon: "🏭", label: "Склад (управление)", allowedRoles: ["admin", "manager"] },
  { to: "/analytics", icon: "📈", label: "Аналитика", allowedRoles: ["admin", "manager"] },
  { to: "/settings", icon: "⚙️", label: "Настройки", allowedRoles: ["admin"] },
];

export default function Layout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const visibleNav = NAV.filter(
    (item) => !item.allowedRoles || item.allowedRoles.includes(user?.role || "")
  );

  return (
    <div className="flex h-screen bg-gray-50 overflow-hidden">
      {/* ── Sidebar ──────────────────────────────────────── */}
      <aside className="w-56 shrink-0 bg-white border-r border-gray-100 flex flex-col">
        {/* Logo */}
        <div className="px-5 py-4 border-b border-gray-100">
          <div className="flex items-center gap-2.5">
            <span className="text-2xl leading-none">☕</span>
            <div>
              <p className="font-bold text-brand text-sm leading-tight">Coffeeshop</p>
              <p className="text-[11px] text-gray-400 leading-tight">ERP v1.0</p>
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-3 py-3 space-y-0.5 overflow-y-auto">
          {visibleNav.map(({ to, icon, label }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                [
                  "flex items-center gap-2.5 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors",
                  isActive
                    ? "bg-brand-50 text-brand"
                    : "text-gray-500 hover:bg-gray-50 hover:text-gray-800",
                ].join(" ")
              }
            >
              <span className="text-base leading-none">{icon}</span>
              {label}
            </NavLink>
          ))}
        </nav>

        {/* User + logout */}
        <div className="px-4 py-4 border-t border-gray-100 space-y-3">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-full bg-brand-100 text-brand font-bold text-sm flex items-center justify-center shrink-0">
              {user?.full_name?.[0]?.toUpperCase() ?? "?"}
            </div>
            <div className="min-w-0">
              <p className="text-sm font-medium text-gray-900 truncate leading-tight">
                {user?.full_name}
              </p>
              <p className="text-[11px] text-gray-400 leading-tight capitalize">{user?.role}</p>
            </div>
          </div>
          <button
            onClick={() => { logout(); navigate("/login"); }}
            className="text-xs text-gray-400 hover:text-red-500 transition-colors"
          >
            Выйти из системы
          </button>
        </div>
      </aside>

      {/* ── Main content ─────────────────────────────────── */}
      <main className="flex-1 overflow-y-auto">{children}</main>
    </div>
  );
}
