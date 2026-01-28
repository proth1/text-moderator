import { useEffect } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAuth } from "./hooks/useAuth";
import MainLayout from "./components/layout/MainLayout";
import Dashboard from "./pages/Dashboard";
import ModerationDemo from "./pages/ModerationDemo";
import ModeratorQueue from "./pages/ModeratorQueue";
import ReviewDetail from "./pages/ReviewDetail";
import PolicyList from "./pages/PolicyList";
import PolicyEditor from "./pages/PolicyEditor";
import AuditLog from "./pages/AuditLog";
import Analytics from "./pages/Analytics";
import Login from "./pages/Login";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
});

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
}

function AppRoutes() {
  const { checkAuth, isAuthenticated, login } = useAuth();

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  // Auto-login for demo mode only — gated behind VITE_DEMO_MODE env var.
  useEffect(() => {
    if (import.meta.env.VITE_DEMO_MODE === "true" && !isAuthenticated) {
      login("tk_admin_test_key_001", {
        id: "a0000000-0000-0000-0000-000000000001",
        email: "admin@civitas.test",
        role: "admin" as import("./types").UserRole,
        created_at: new Date().toISOString(),
      });
    }
  }, [isAuthenticated, login]);

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <PrivateRoute>
            <MainLayout />
          </PrivateRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="demo" element={<ModerationDemo />} />
        <Route path="reviews" element={<ModeratorQueue />} />
        <Route path="reviews/:id" element={<ReviewDetail />} />
        <Route path="policies" element={<PolicyList />} />
        <Route path="policies/new" element={<PolicyEditor />} />
        <Route path="policies/:id/edit" element={<PolicyEditor />} />
        <Route path="audit" element={<AuditLog />} />
        <Route path="analytics" element={<Analytics />} />
      </Route>
    </Routes>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
