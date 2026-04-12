import { formatDateTime } from "../lib/formatters";

export function DashboardSummaryBlock({ summary }) {
  const summaryText =
    summary.summary || "Сводка пока не сформирована. Для демонстрации используются доступные backend-данные.";

  return (
    <section className="card summary-card">
      <div className="summary-copy">
        <p className="page-eyebrow">Сводка</p>
        <h2 className="section-title section-title-lg">Общая сводка по рынку и сигналам</h2>
        <p className="hero-text">{summaryText}</p>
      </div>

      <dl className="summary-stats">
        <div className="summary-stat">
          <dt>Обновлено</dt>
          <dd>{formatDateTime(summary.generated_at)}</dd>
        </div>
        <div className="summary-stat">
          <dt>Активов в сводке</dt>
          <dd>{summary.assets?.length ?? 0}</dd>
        </div>
        <div className="summary-stat">
          <dt>Последних прогнозов</dt>
          <dd>{summary.latest_forecasts?.length ?? 0}</dd>
        </div>
        <div className="summary-stat">
          <dt>Событий в блоке</dt>
          <dd>{summary.recent_events?.length ?? 0}</dd>
        </div>
      </dl>
    </section>
  );
}
