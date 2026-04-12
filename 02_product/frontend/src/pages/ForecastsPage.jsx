import { useState } from "react";
import { api } from "../api/client";
import { EmptyState } from "../components/EmptyState";
import { LoadingBlock } from "../components/LoadingBlock";
import { PageHeader } from "../components/PageHeader";
import { RegimeBlock } from "../components/RegimeBlock";
import {
  formatDateTime,
  formatPercent,
  getConfidenceLabel,
  getDirectionClassName,
  getDirectionLabel,
  getStrengthLabel
} from "../lib/formatters";
import { useRemoteData } from "../lib/useRemoteData";

export function ForecastsPage() {
  const [query, setQuery] = useState("");
  const [directionFilter, setDirectionFilter] = useState("all");
  const [selectedForecastId, setSelectedForecastId] = useState(null);

  const { data, loading, error, reload } = useRemoteData(
    async () => {
      const [dashboardSummary, latestForecast] = await Promise.all([
        api.getDashboardSummary(),
        api.getLatestForecast()
      ]);

      return {
        latestForecasts: dashboardSummary.latest_forecasts || [],
        latestSingleForecast: latestForecast || null,
        regime: dashboardSummary.regime || null
      };
    },
    []
  );

  if (loading) {
    return <LoadingBlock label="Загрузка прогнозов..." />;
  }

  if (error) {
    return <EmptyState title="Не удалось загрузить прогнозы" description={error} />;
  }

  const allForecasts = data.latestSingleForecast
    ? [
        data.latestSingleForecast,
        ...(data.latestForecasts || []).filter((forecast) => forecast.id !== data.latestSingleForecast.id)
      ]
    : data.latestForecasts || [];

  const filteredForecasts = allForecasts.filter((forecast) => {
    const matchesDirection = directionFilter === "all" || forecast.direction === directionFilter;
    const haystack = `${forecast.asset_ticker} ${forecast.asset_name} ${forecast.explanation}`.toLowerCase();
    const matchesQuery = !query.trim() || haystack.includes(query.trim().toLowerCase());
    return matchesDirection && matchesQuery;
  });

  const selectedForecast =
    filteredForecasts.find((forecast) => forecast.id === selectedForecastId) || filteredForecasts[0] || null;
  const averageConfidence =
    filteredForecasts.length > 0
      ? filteredForecasts.reduce((sum, forecast) => sum + Number(forecast.confidence || 0), 0) / filteredForecasts.length
      : null;

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Forecasts"
        title="Прогнозы"
        description="Лента последних сигналов, их силы, уверенности и краткого объяснения без выдумывания новой бизнес-логики."
        actions={
          <button className="secondary-button" onClick={reload} type="button">
            Обновить
          </button>
        }
      />

      <section className="kpi-grid">
        <article className="kpi-card">
          <span className="kpi-label">Bullish forecasts</span>
          <strong className="kpi-value">
            {allForecasts.filter((forecast) => forecast.direction === "positive").length}
          </strong>
          <span className="kpi-meta">направление с ростом</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Bearish forecasts</span>
          <strong className="kpi-value">
            {allForecasts.filter((forecast) => forecast.direction === "negative").length}
          </strong>
          <span className="kpi-meta">сигналы со снижением</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Average confidence</span>
          <strong className="kpi-value">
            {averageConfidence === null ? "-" : formatPercent(averageConfidence, { multiplier: 100 })}
          </strong>
          <span className="kpi-meta">по текущей фильтрации</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Latest generation</span>
          <strong className="kpi-value">
            {selectedForecast ? formatDateTime(selectedForecast.generated_at) : "-"}
          </strong>
          <span className="kpi-meta">самый свежий расчет</span>
        </article>
      </section>

      <section className="toolbar-card card">
        <label className="search-field">
          <span className="search-icon" aria-hidden="true">
            ⌕
          </span>
          <input
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Filter by asset"
            type="search"
            value={query}
          />
        </label>

        <div className="chip-group">
          {[
            { label: "All", value: "all" },
            { label: "Bullish", value: "positive" },
            { label: "Bearish", value: "negative" },
            { label: "Neutral", value: "neutral" }
          ].map((item) => (
            <button
              key={item.value}
              className={`filter-chip${directionFilter === item.value ? " filter-chip-active" : ""}`}
              onClick={() => setDirectionFilter(item.value)}
              type="button"
            >
              {item.label}
            </button>
          ))}
        </div>
      </section>

      <RegimeBlock compact regime={data.regime} />

      {filteredForecasts.length === 0 ? (
        <EmptyState
          description="Попробуйте сбросить фильтры или изменить поисковый запрос."
          title="По выбранным условиям прогнозы не найдены"
        />
      ) : (
        <>
          <section className="card forecast-table-card">
            <div className="forecast-table-header">
              <span>Asset</span>
              <span>Direction</span>
              <span>Strength</span>
              <span>Confidence</span>
              <span>Explanation preview</span>
              <span>Created</span>
              <span>Action</span>
            </div>

            <div className="forecast-table-body">
              {filteredForecasts.map((forecast) => (
                <button
                  className={`forecast-row${selectedForecast?.id === forecast.id ? " forecast-row-active" : ""}`}
                  key={forecast.id}
                  onClick={() => setSelectedForecastId(forecast.id)}
                  type="button"
                >
                  <span className="forecast-row-asset">{forecast.asset_ticker}</span>
                  <span
                    className={`forecast-row-direction forecast-row-direction-${getDirectionClassName(
                      forecast.direction
                    )}`}
                  >
                    {getDirectionLabel(forecast.direction)}
                  </span>
                  <span>{getStrengthLabel(forecast.strength)}</span>
                  <span>{getConfidenceLabel(forecast.confidence)}</span>
                  <span className="forecast-row-text">{forecast.explanation}</span>
                  <span>{formatDateTime(forecast.generated_at)}</span>
                  <span className="text-link">Открыть</span>
                </button>
              ))}
            </div>
          </section>

          {selectedForecast ? (
            <section className="card">
              <div className="section-header">
                <div>
                  <h2 className="section-title">
                    {selectedForecast.asset_ticker} · {getDirectionLabel(selectedForecast.direction)}
                  </h2>
                  <p className="section-subtitle">{formatDateTime(selectedForecast.generated_at)}</p>
                </div>
                <span className={`badge badge-${getDirectionClassName(selectedForecast.direction)}`}>
                  {getConfidenceLabel(selectedForecast.confidence)} confidence
                </span>
              </div>

              <p>{selectedForecast.explanation}</p>

              <div className="tag-row">
                <span className="badge badge-muted">
                  Strength {formatPercent(selectedForecast.strength, { multiplier: 100 })}
                </span>
                <span className="badge badge-muted">
                  Confidence {formatPercent(selectedForecast.confidence, { multiplier: 100 })}
                </span>
                <span className="badge badge-muted">{selectedForecast.horizon}</span>
                {selectedForecast.market_context_label ? (
                  <span className="badge badge-muted">
                    {selectedForecast.market_context_label} ·{" "}
                    {formatPercent(selectedForecast.market_context_score, { multiplier: 100 })}
                  </span>
                ) : null}
              </div>
            </section>
          ) : null}
        </>
      )}
    </div>
  );
}
