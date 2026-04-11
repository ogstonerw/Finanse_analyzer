import { formatDateTime, formatPercent, getDirectionLabel } from "../lib/formatters";

export function ForecastsBlock({ forecasts, title = "Прогнозы" }) {
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
              <span className={`badge badge-${forecast.direction}`}>{getDirectionLabel(forecast.direction)}</span>
            </div>

            <dl className="inline-metrics">
              <div>
                <dt>Горизонт</dt>
                <dd>{forecast.horizon}</dd>
              </div>
              <div>
                <dt>Сила</dt>
                <dd>{formatPercent(forecast.strength, { multiplier: 100 })}</dd>
              </div>
              <div>
                <dt>Уверенность</dt>
                <dd>{formatPercent(forecast.confidence, { multiplier: 100 })}</dd>
              </div>
              <div>
                <dt>Сгенерирован</dt>
                <dd>{formatDateTime(forecast.generated_at)}</dd>
              </div>
            </dl>

            <p>{forecast.explanation}</p>

            {forecast.market_context ? (
              <p className="muted-text">
                Контекст рынка: {forecast.market_context.label} ·{" "}
                {formatPercent(forecast.market_context.score, { multiplier: 100 })}
              </p>
            ) : null}

            {!forecast.market_context && forecast.market_context_label ? (
              <p className="muted-text">
                Контекст рынка: {forecast.market_context_label} ·{" "}
                {formatPercent(forecast.market_context_score, { multiplier: 100 })}
              </p>
            ) : null}
          </article>
        ))}
      </div>
    </section>
  );
}
