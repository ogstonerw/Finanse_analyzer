import { formatDateTime } from "../lib/formatters";

export function DashboardSummaryBlock({ summary }) {
  return (
    <section className="card hero-card">
      <div className="hero-copy">
        <p className="eyebrow">Dashboard Summary</p>
        <h2 className="hero-title">Общая сводка по рынку и последним сигналам</h2>
        <p className="hero-text">{summary.summary}</p>
      </div>

      <dl className="hero-stats">
        <div>
          <dt>Обновлено</dt>
          <dd>{formatDateTime(summary.generated_at)}</dd>
        </div>
        <div>
          <dt>Активов в сводке</dt>
          <dd>{summary.assets.length}</dd>
        </div>
        <div>
          <dt>Последних прогнозов</dt>
          <dd>{summary.latest_forecasts.length}</dd>
        </div>
        <div>
          <dt>Событий в блоке</dt>
          <dd>{summary.recent_events.length}</dd>
        </div>
      </dl>
    </section>
  );
}
