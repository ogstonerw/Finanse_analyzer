import { useState } from "react";
import { api } from "../api/client";
import { EmptyState } from "../components/EmptyState";
import { LoadingBlock } from "../components/LoadingBlock";
import { NoticeCard } from "../components/NoticeCard";
import { PageHeader } from "../components/PageHeader";
import { RegimeBlock } from "../components/RegimeBlock";
import {
  formatDateTime,
  formatPercent,
  getConfidenceLabel,
  getDirectionClassName,
  getDirectionLabel,
  getStrengthLabel,
  normalizeForecastDirection
} from "../lib/formatters";
import { useRemoteData } from "../lib/useRemoteData";

export function ForecastsPage() {
  const [query, setQuery] = useState("");
  const [directionFilter, setDirectionFilter] = useState("all");
  const [generateError, setGenerateError] = useState("");
  const [generateEventId, setGenerateEventId] = useState("");
  const [generateNotice, setGenerateNotice] = useState("");
  const [generateTicker, setGenerateTicker] = useState("");
  const [generating, setGenerating] = useState(false);
  const [selectedForecastId, setSelectedForecastId] = useState(null);

  const { data, loading, error, reload } = useRemoteData(
    async () => {
      const [dashboardResult, latestResult] = await Promise.allSettled([
        api.getDashboardSummary(),
        api.getLatestForecast()
      ]);

      if (dashboardResult.status === "rejected" && latestResult.status === "rejected") {
        throw new Error("Backend не вернул прогнозы ни из dashboard summary, ни из latest forecast endpoint.");
      }

      const noticeParts = [];

      if (dashboardResult.status === "rejected") {
        noticeParts.push("Dashboard summary временно недоступен, поэтому показан только latest forecast.");
      }

      if (latestResult.status === "rejected") {
        noticeParts.push("Отдельный latest forecast endpoint временно недоступен.");
      }

      const dashboardSummary =
        dashboardResult.status === "fulfilled"
          ? dashboardResult.value
          : { latest_forecasts: [], regime: null };

      return {
        assets: dashboardSummary.assets || [],
        latestForecasts: dashboardSummary.latest_forecasts || [],
        latestSingleForecast: latestResult.status === "fulfilled" ? latestResult.value || null : null,
        notice: noticeParts.join(" "),
        recentEvents: dashboardSummary.recent_events || [],
        regime: dashboardSummary.regime || null
      };
    },
    []
  );

  if (loading) {
    return <LoadingBlock label="Загрузка прогнозов..." />;
  }

  if (error) {
    return (
      <EmptyState
        action={
          <button className="secondary-button" onClick={reload} type="button">
            Повторить запрос
          </button>
        }
        description={error}
        title="Не удалось загрузить прогнозы"
        variant="error"
      />
    );
  }

  const allForecasts = data.latestSingleForecast
    ? [
        data.latestSingleForecast,
        ...(data.latestForecasts || []).filter((forecast) => forecast.id !== data.latestSingleForecast.id)
      ]
    : data.latestForecasts || [];

  const filteredForecasts = allForecasts.filter((forecast) => {
    const direction = normalizeForecastDirection(forecast.direction);
    const matchesDirection = directionFilter === "all" || direction === directionFilter;
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
  const selectedGenerateTicker = generateTicker || data.assets[0]?.ticker || "";

  async function handleGenerateForecast(event) {
    event.preventDefault();
    if (!selectedGenerateTicker || generating) {
      return;
    }

    setGenerating(true);
    setGenerateError("");
    setGenerateNotice("");

    try {
      const forecast = await api.generateForecast({
        ticker: selectedGenerateTicker,
        event_id: generateEventId || undefined,
        horizon: "1w"
      });
      setSelectedForecastId(forecast.id);
      setGenerateNotice(`Прогноз для ${forecast.asset_ticker} сформирован и сохранен в истории.`);
      await reload();
    } catch (err) {
      setGenerateError(err instanceof Error ? err.message : "Не удалось сформировать прогноз");
    } finally {
      setGenerating(false);
    }
  }

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Сигналы"
        title="Прогнозы"
        description="Лента последних сигналов, их силы, уверенности и краткого объяснения без выдумывания новой бизнес-логики."
        actions={
          <button className="secondary-button" onClick={reload} type="button">
            Обновить
          </button>
        }
      />

      {data.notice ? <NoticeCard description={data.notice} title="Частичная деградация данных" /> : null}
      {generateNotice ? <NoticeCard description={generateNotice} title="Прогноз сформирован" /> : null}
      {generateError ? <NoticeCard description={generateError} title="Ошибка генерации прогноза" variant="warning" /> : null}

      <section className="toolbar-card card">
        <form className="forecast-generate-form" onSubmit={handleGenerateForecast}>
          <label className="form-field">
            <span>Актив</span>
            <select
              onChange={(event) => setGenerateTicker(event.target.value)}
              value={selectedGenerateTicker}
            >
              {(data.assets || []).map((asset) => (
                <option key={asset.id} value={asset.ticker}>
                  {asset.ticker} · {asset.name}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field form-field-wide">
            <span>Событие</span>
            <select onChange={(event) => setGenerateEventId(event.target.value)} value={generateEventId}>
              <option value="">Автоматически: последнее релевантное событие</option>
              {(data.recentEvents || []).map((item) => (
                <option key={item.id} value={item.id}>
                  {item.asset_ticker ? `${item.asset_ticker} · ` : ""}
                  {item.event_type} · {item.news_title}
                </option>
              ))}
            </select>
          </label>

          <button className="primary-button" disabled={!selectedGenerateTicker || generating} type="submit">
            {generating ? "Формируется..." : "Сформировать прогноз"}
          </button>
        </form>
      </section>

      <section className="kpi-grid">
        <article className="kpi-card">
          <span className="kpi-label">Сигналы на рост</span>
          <strong className="kpi-value">
            {allForecasts.filter((forecast) => normalizeForecastDirection(forecast.direction) === "up").length}
          </strong>
          <span className="kpi-meta">направление с ростом</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Сигналы на снижение</span>
          <strong className="kpi-value">
            {allForecasts.filter((forecast) => normalizeForecastDirection(forecast.direction) === "down").length}
          </strong>
          <span className="kpi-meta">сигналы со снижением</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Средняя уверенность</span>
          <strong className="kpi-value">
            {averageConfidence === null ? "-" : formatPercent(averageConfidence, { multiplier: 100 })}
          </strong>
          <span className="kpi-meta">по текущей фильтрации</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Последний расчет</span>
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
            placeholder="Поиск по тикеру или названию актива"
            type="search"
            value={query}
          />
        </label>

        <div className="chip-group">
          {[
            { label: "Все", value: "all" },
            { label: "Рост", value: "up" },
            { label: "Снижение", value: "down" },
            { label: "Нейтрально", value: "neutral" }
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
          action={
            <button
              className="secondary-button"
              onClick={() => {
                setDirectionFilter("all");
                setQuery("");
              }}
              type="button"
            >
              Сбросить фильтры
            </button>
          }
          description="Попробуйте сбросить фильтры или изменить поисковый запрос."
          title="По выбранным условиям прогнозы не найдены"
        />
      ) : (
        <>
          <section className="card forecast-table-card">
            <div className="forecast-table-header">
              <span>Актив</span>
              <span>Направление</span>
              <span>Сила</span>
              <span>Уверенность</span>
              <span>Краткое объяснение</span>
              <span>Создан</span>
              <span>Действие</span>
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
                  <span className="text-link">Выбрать</span>
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
                  Уверенность: {getConfidenceLabel(selectedForecast.confidence)}
                </span>
              </div>

              <p>{selectedForecast.explanation}</p>

              <div className="tag-row">
                <span className="badge badge-muted">
                  Сила: {formatPercent(selectedForecast.strength, { multiplier: 100 })}
                </span>
                <span className="badge badge-muted">
                  Уверенность: {formatPercent(selectedForecast.confidence, { multiplier: 100 })}
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
