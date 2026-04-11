import { NavLink, Outlet } from "react-router-dom";

const navItems = [
  { label: "Дашборд", to: "/dashboard" },
  { label: "Активы", to: "/assets" },
  { label: "Новости и события", to: "/news" },
  { label: "Прогнозы", to: "/forecasts" }
];

export function AppLayout({ session, onLogout }) {
  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Market AI Platform</p>
          <h1 className="app-title">MVP мониторинга рынка и прогнозов</h1>
        </div>

        <div className="topbar-actions">
          <div className="session-badge">
            <span className="session-label">Сессия</span>
            <span className="session-value">{session?.email}</span>
          </div>
          <button className="secondary-button" onClick={onLogout} type="button">
            Выйти
          </button>
        </div>
      </header>

      <div className="layout-grid">
        <aside className="sidebar">
          <nav aria-label="Основная навигация" className="nav-list">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                className={({ isActive }) => (isActive ? "nav-link nav-link-active" : "nav-link")}
                to={item.to}
              >
                {item.label}
              </NavLink>
            ))}
          </nav>

          <div className="sidebar-note">
            <h2>Контур MVP</h2>
            <p>
              Интерфейс показывает ключевые сущности платформы: режим рынка, активы, новости,
              события и последние прогнозы.
            </p>
          </div>
        </aside>

        <main className="page-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
