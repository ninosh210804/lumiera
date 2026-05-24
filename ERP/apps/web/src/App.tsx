import { Routes, Route, Navigate } from "react-router-dom";
import { useAuth } from "./lib/auth";
import Layout from "./components/Layout";
import LoginPage from "./pages/LoginPage";
import DashboardPage from "./pages/DashboardPage";
import OrdersPage from "./pages/OrdersPage";
import POSPage from "./pages/POSPage";
import MenuPage from "./pages/MenuPage";
import InventoryPage from "./pages/InventoryPage";
import WarehousePage from "./pages/WarehousePage";
import SalesPage from "./pages/SalesPage";
import AnalyticsPage from "./pages/AnalyticsPage";
import SettingsPage from "./pages/SettingsPage";

function RoleProtectedRoute({
  children,
  allowedRoles,
}: {
  children: React.ReactNode;
  allowedRoles: string[];
}) {
  const { user, token } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  if (!user || !allowedRoles.includes(user.role)) {
    const fallback = user?.role === "barista" ? "/pos" : "/dashboard";
    return <Navigate to={fallback} replace />;
  }
  return <Layout>{children}</Layout>;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<RoleAwareFallback />} />

      <Route path="/dashboard" element={<RoleProtectedRoute allowedRoles={["admin", "manager"]}><DashboardPage /></RoleProtectedRoute>} />
      <Route path="/orders"    element={<RoleProtectedRoute allowedRoles={["admin", "manager", "barista"]}><OrdersPage /></RoleProtectedRoute>} />
      <Route path="/pos"       element={<RoleProtectedRoute allowedRoles={["barista", "manager", "admin"]}><POSPage /></RoleProtectedRoute>} />
      <Route path="/menu"      element={<RoleProtectedRoute allowedRoles={["admin", "manager"]}><MenuPage /></RoleProtectedRoute>} />
      <Route path="/inventory" element={<RoleProtectedRoute allowedRoles={["admin", "manager"]}><InventoryPage /></RoleProtectedRoute>} />
      <Route path="/warehouse" element={<RoleProtectedRoute allowedRoles={["admin", "manager"]}><WarehousePage /></RoleProtectedRoute>} />
      <Route path="/sales"     element={<RoleProtectedRoute allowedRoles={["admin", "manager"]}><SalesPage /></RoleProtectedRoute>} />
      <Route path="/analytics" element={<RoleProtectedRoute allowedRoles={["admin", "manager"]}><AnalyticsPage /></RoleProtectedRoute>} />
      <Route path="/settings"  element={<RoleProtectedRoute allowedRoles={["admin"]}><SettingsPage /></RoleProtectedRoute>} />

      <Route path="*" element={<RoleAwareFallback />} />
    </Routes>
  );
}

function RoleAwareFallback() {
  const { user, token } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  // Baristas land on POS; everyone else on the dashboard.
  const target = user?.role === "barista" ? "/pos" : "/dashboard";
  return <Navigate to={target} replace />;
}
