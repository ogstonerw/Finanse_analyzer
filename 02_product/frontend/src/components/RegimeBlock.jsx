import { formatDateTime, formatPercent, getRegimeLabel } from "../lib/formatters";

function Metric({ label, value }) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{formatPercent(value, { multiplier: 100 })}</strong>
    </div>
  );
}

export function RegimeBlock({ regime }) {
  if (!regime) {
    return null;
  }

  const subScores = regime.sub_scores || {};

  return (
    <section className="card regime-card">
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

      <p className="card-lead">{regime.summary}</p>
      <p>{regime.explanation}</p>

      <div className="metrics-grid">
        <Metric label="Market stress" value={subScores.market_stress} />
        <Metric label="News stress" value={subScores.news_stress} />
        <Metric label="Macro stress" value={subScores.macro_stress} />
        <Metric label="Commodity stress" value={subScores.commodity_stress} />
        <Metric label="Breadth stress" value={subScores.breadth_stress} />
      </div>
    </section>
  );
}
