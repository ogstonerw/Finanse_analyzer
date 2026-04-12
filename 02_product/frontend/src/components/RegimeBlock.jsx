import { formatDateTime, formatPercent, getRegimeLabel } from "../lib/formatters";

function Metric({ label, value }) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{formatPercent(value, { multiplier: 100 })}</strong>
    </div>
  );
}

export function RegimeBlock({ compact = false, regime }) {
  if (!regime) {
    return (
      <section className={`card regime-card regime-card-empty${compact ? " regime-card-compact" : ""}`}>
        <div className="section-header">
          <div>
            <h2 className="section-title">Кризисометр и режим рынка</h2>
            <p className="section-subtitle">Данные для расчета режима пока недоступны</p>
          </div>
        </div>

        <div className="inline-empty">
          Backend пока не вернул рассчитанный режим рынка. Остальные блоки MVP можно демонстрировать отдельно.
        </div>
      </section>
    );
  }

  const subScores = regime.sub_scores || {};
  const regimeWidth = Math.max(8, Math.min(100, Number(regime.regime_score || 0) * 100));

  return (
    <section className={`card regime-card${compact ? " regime-card-compact" : ""}`}>
      <div className="section-header">
        <div>
          <h2 className="section-title">Кризисометр и режим рынка</h2>
          <p className="section-subtitle">
            {formatDateTime(regime.calculated_at)} · {regime.calculation_model}
          </p>
        </div>
        <span className={`badge badge-regime-${regime.regime_label}`}>
          {getRegimeLabel(regime.regime_label)}
        </span>
      </div>

      <div className="regime-score">
        <span>Интегральный score</span>
        <strong>{formatPercent(regime.regime_score, { multiplier: 100 })}</strong>
      </div>

      <div className="progress-track" aria-hidden="true">
        <span className="progress-bar" style={{ width: `${regimeWidth}%` }} />
      </div>

      <p className="card-lead">{regime.summary || "Режим рынка рассчитан без расширенного summary."}</p>
      <p>{regime.explanation || "Подробное объяснение режима пока не сформировано."}</p>

      <div className="metrics-grid">
        <Metric label="Рыночный стресс" value={subScores.market_stress} />
        <Metric label="Новостный стресс" value={subScores.news_stress} />
        <Metric label="Макро-стресс" value={subScores.macro_stress} />
        <Metric label="Сырьевой стресс" value={subScores.commodity_stress} />
        <Metric label="Ширина рынка" value={subScores.breadth_stress} />
      </div>
    </section>
  );
}
