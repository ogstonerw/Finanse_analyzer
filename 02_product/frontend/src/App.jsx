import { Navigate, Route, Routes } from "react-router-dom";
import { useState } from "react";
import { loadSession, clearSession, saveSession } from "./lib/authStorage";
import { AppLayout } from "./components/AppLayout";
import { LoginPage } from "./pages/LoginPage";
import { DashboardPage } from "./pages/DashboardPage";
import { AssetsPage } from "./pages/AssetsPage";
import { AssetDetailsPage } from "./pages/AssetDetailsPage";
import { NewsEventsPage } from "./pages/NewsEventsPage";
import { ForecastsPage } from "./pages/ForecastsPage";

function RequireAuth({ session, children }) {
  if (!session) {
    return <Navigate replace to="/login" />;
  }

  return children;
}

export default function App() {
  const [session, setSession] = useState(() => loadSession());

  function handleLogin(nextSession) {
    saveSession(nextSession);
    setSession(nextSession);
  }

  function handleLogout() {
    clearSession();
    setSession(null);
  }

  return (
    <Routes>
      <Route
        element={session ? <Navigate replace to="/dashboard" /> : <LoginPage onLogin={handleLogin} />}
        path="/login"
      />

      <Route
        element={
          <RequireAuth session={session}>
            <AppLayout onLogout={handleLogout} session={session} />
          </RequireAuth>
        }
        path="/"
      >
        <Route element={<Navigate replace to="/dashboard" />} index />
        <Route element={<DashboardPage />} path="dashboard" />
        <Route element={<AssetsPage />} path="assets" />
        <Route element={<AssetDetailsPage />} path="assets/:ticker" />
        <Route element={<NewsEventsPage />} path="news" />
        <Route element={<ForecastsPage />} path="forecasts" />
      </Route>

      <Route path="*" element={<Navigate replace to={session ? "/dashboard" : "/login"} />} />
    </Routes>
  );
}
