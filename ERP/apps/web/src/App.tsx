import { Routes, Route, Navigate } from "react-router-dom";
import { useAuth } from "./lib/auth";
import Layout from "./components/Layout";
import LoginPage from "./pages/LoginPage";
import DashboardPage from "./pages/DashboardPage";
import OrdersPage from "./pages/OrdersPage";
import POSPage from "./pages/POSPage";
import MenuPage from "./pages/MenuPage";
import ClientsPage from "./pages/ClientsPage";
import DeliveryPage from "./pages/DeliveryPage";
import OnlineOrdersPage from "./pages/OnlineOrdersPage";
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
      {/* Public customer delivery app (opened via QR: /order?location=<uuid>) */}
      <Route path="/order" element={<DeliveryPage />} />
      <Route path="/" element={<RoleAwareFallback />} />

      <Route
        path="/dashboard"
        element={
          <RoleProtectedRoute allowedRoles={["admin", "manager"]}>
            <DashboardPage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/orders"
        element={
          <RoleProtectedRoute allowedRoles={["admin", "manager", "barista"]}>
            <OrdersPage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/pos"
        element={
          <RoleProtectedRoute allowedRoles={["barista", "manager", "admin"]}>
            <POSPage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/menu"
        element={
          <RoleProtectedRoute allowedRoles={["admin", "manager"]}>
            <MenuPage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/clients"
        element={
          <RoleProtectedRoute allowedRoles={["admin", "manager", "barista"]}>
            <ClientsPage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/online-orders"
        element={
          <RoleProtectedRoute allowedRoles={["admin", "manager", "barista"]}>
            <OnlineOrdersPage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/inventory"
        element={
          <RoleProtectedRoute allowedRoles={["admin", "manager", "barista"]}>
            <InventoryPage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/warehouse"
        element={
          <RoleProtectedRoute allowedRoles={["admin", "manager", "barista"]}>
            <WarehousePage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/sales"
        element={
          <RoleProtectedRoute allowedRoles={["admin", "manager"]}>
            <SalesPage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/analytics"
        element={
          <RoleProtectedRoute allowedRoles={["admin", "manager"]}>
            <AnalyticsPage />
          </RoleProtectedRoute>
        }
      />
      <Route
        path="/settings"
        element={
          <RoleProtectedRoute allowedRoles={["admin"]}>
            <SettingsPage />
          </RoleProtectedRoute>
        }
      />

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
