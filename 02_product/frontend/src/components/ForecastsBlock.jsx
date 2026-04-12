import {
  formatDateTime,
  formatPercent,
  getConfidenceLabel,
  getDirectionClassName,
  getDirectionLabel,
  getStrengthLabel
} from "../lib/formatters";
import { EmptyState } from "./EmptyState";

export function ForecastsBlock({ forecasts, title = "Прогнозы" }) {
  if (!forecasts?.length) {
    return (
      <section className="card">
        <div className="section-header">
          <h2 className="section-title">{title}</h2>
          <span className="badge badge-muted">0</span>
        </div>

        <EmptyState
          description="Для этого блока backend пока не вернул актуальных прогнозов."
          title="Прогнозов пока нет"
        />
      </section>
    );
  }

  return (
    <section className="card">
      <div className="section-header">
        <h2 className="section-title">{title}</h2>
        <span className="badge badge-muted">{forecasts.length}</span>
      </div>

      <div className="stack-list">
        {forecasts.map((forecast) => (
          <article className="list-item" key={forecast.id}>
            <div className="list-item-head">
              <h3>
                {forecast.asset_ticker} · {forecast.asset_name}
              </h3>
              <span className={`badge badge-${getDirectionClassName(forecast.direction)}`}>
                {getDirectionLabel(forecast.direction)}
              </span>
            </div>

            <dl className="inline-metrics">
              <div>
                <dt>Горизонт</dt>
                <dd>{forecast.horizon}</dd>
              </div>
              <div>
                <dt>Сила</dt>
                <dd>{getStrengthLabel(forecast.strength)}</dd>
              </div>
              <div>
                <dt>Уверенность</dt>
                <dd>{getConfidenceLabel(forecast.confidence)}</dd>
              </div>
              <div>
                <dt>Сгенерирован</dt>
                <dd>{formatDateTime(forecast.generated_at)}</dd>
              </div>
            </dl>

            <p>{forecast.explanation}</p>

            <div className="tag-row">
              <span className="badge badge-muted">
                Сила: {formatPercent(forecast.strength, { multiplier: 100 })}
              </span>
              <span className="badge badge-muted">
                Уверенность: {formatPercent(forecast.confidence, { multiplier: 100 })}
              </span>
              {forecast.market_context ? (
                <span className="badge badge-muted">
                  {forecast.market_context.label} ·{" "}
                  {formatPercent(forecast.market_context.score, { multiplier: 100 })}
                </span>
              ) : null}
              {!forecast.market_context && forecast.market_context_label ? (
                <span className="badge badge-muted">
                  {forecast.market_context_label} ·{" "}
                  {formatPercent(forecast.market_context_score, { multiplier: 100 })}
                </span>
              ) : null}
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
