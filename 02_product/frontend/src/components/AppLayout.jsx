import { NavLink, Outlet } from "react-router-dom";

const navItems = [
  { label: "Панель обзора", to: "/dashboard" },
  { label: "Активы", to: "/assets" },
  { label: "Новости и события", to: "/news" },
  { label: "Прогнозы", to: "/forecasts" }
];

export function AppLayout({ session, onLogout }) {
  return (
    <div className="app-shell">
      <aside className="sidebar-shell">
        <div className="brand-block">
          <p className="brand-title">MRAP</p>
          <p className="brand-subtitle">Market Reaction Analytics Platform</p>
        </div>

        <div className="sidebar">
          <nav aria-label="Основная навигация" className="nav-list">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                className={({ isActive }) => (isActive ? "nav-link nav-link-active" : "nav-link")}
                to={item.to}
              >
                <span className="nav-link-dot" aria-hidden="true" />
                {item.label}
              </NavLink>
            ))}
          </nav>

          <div className="sidebar-panel">
            <h2 className="sidebar-panel-title">Контур MVP</h2>
            <p>
              Интерфейс показывает ключевые сущности платформы: режим рынка, активы, новости,
              события и последние прогнозы поверх уже существующего backend API.
            </p>
          </div>

          <div className="sidebar-user-card">
            <div className="avatar-badge" aria-hidden="true">
              {session?.email?.slice(0, 1)?.toUpperCase() || "U"}
            </div>
            <div className="sidebar-user-copy">
              <strong>{session?.email || "demo@local"}</strong>
              <span>Активная сессия</span>
            </div>
          </div>
        </div>
      </aside>

      <div className="workspace-shell">
        <header className="topbar">
          <label className="topbar-search" aria-label="Глобальный поиск по интерфейсу">
            <span className="search-icon" aria-hidden="true">
              ⌕
            </span>
            <input placeholder="Демо-поиск по активам, новостям и прогнозам" type="search" />
          </label>

          <div className="topbar-meta">
            <span className="status-badge status-badge-positive">MVP готов к демонстрации</span>
            <div className="session-chip">
              <span className="session-chip-label">Сессия</span>
              <strong>{session?.email}</strong>
            </div>
            <button className="secondary-button" onClick={onLogout} type="button">
              Выйти
            </button>
          </div>
        </header>

        <main className="page-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
