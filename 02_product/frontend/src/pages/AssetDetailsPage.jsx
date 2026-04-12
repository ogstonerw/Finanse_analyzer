import { Link, useParams } from "react-router-dom";
import { api } from "../api/client";
import { EmptyState } from "../components/EmptyState";
import { EventsListBlock } from "../components/EventsListBlock";
import { ForecastsBlock } from "../components/ForecastsBlock";
import { IndicatorsBlock } from "../components/IndicatorsBlock";
import { LoadingBlock } from "../components/LoadingBlock";
import { PageHeader } from "../components/PageHeader";
import { PricesBlock } from "../components/PricesBlock";
import { RegimeBlock } from "../components/RegimeBlock";
import {
  formatDateTime,
  formatPercent,
  formatPrice,
  getAssetTypeLabel,
  getTrendLabel
} from "../lib/formatters";
import { useRemoteData } from "../lib/useRemoteData";

export function AssetDetailsPage() {
  const { ticker = "" } = useParams();
  const normalizedTicker = ticker.toUpperCase();

  const { data, error, loading, reload } = useRemoteData(
    async () => {
      const [asset, prices, indicators, dashboardSummary] = await Promise.all([
        api.getAsset(normalizedTicker),
        api.getAssetPrices(normalizedTicker),
        api.getAssetIndicators(normalizedTicker),
        api.getDashboardSummary()
      ]);

      return {
        asset,
        forecasts: (dashboardSummary.latest_forecasts || []).filter(
          (forecast) => forecast.asset_ticker === normalizedTicker
        ),
        indicators: indicators.items || [],
        prices: prices.items || [],
        relatedEvents: (dashboardSummary.recent_events || []).filter(
          (event) => event.asset_ticker === normalizedTicker
        ),
        regime: dashboardSummary.regime
      };
    },
    [normalizedTicker]
  );

  if (loading) {
    return <LoadingBlock label={`Загрузка деталей ${normalizedTicker}...`} />;
  }

  if (error) {
    return (
      <EmptyState
        description={`${error}. Попробуйте обновить страницу.`}
        title="Не удалось открыть карточку актива"
      />
    );
  }

  const latestIndicator = data.indicators[data.indicators.length - 1] || null;
  const latestPrice = data.prices[data.prices.length - 1] || null;
  const previousPrice = data.prices[data.prices.length - 2] || null;
  const dayChange =
    latestPrice && previousPrice && Number(previousPrice.close_price)
      ? ((Number(latestPrice.close_price) - Number(previousPrice.close_price)) / Number(previousPrice.close_price)) * 100
      : null;
  const latestForecast = data.forecasts[0] || null;
  const dayChangeVariant =
    dayChange === null || dayChange === undefined ? "muted" : dayChange >= 0 ? "positive" : "negative";

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Asset details"
        actions={
          <div className="page-actions">
            <button className="secondary-button" onClick={reload} type="button">
              Обновить
            </button>
            <Link className="secondary-button secondary-link" to="/assets">
              Ко всем активам
            </Link>
          </div>
        }
        description="Карточка инструмента с ценовым рядом, последними индикаторами, сигналом и рыночным контекстом."
        title="Детали актива"
      />

      <section className="card detail-hero-card">
        <div>
          <p className="asset-ticker">{data.asset.ticker}</p>
          <h2 className="hero-title hero-title-compact">{data.asset.name}</h2>
          <p className="hero-text">
            {getAssetTypeLabel(data.asset.asset_type)} · {data.asset.sector || "Sector n/a"} ·{" "}
            {data.asset.currency || "Currency n/a"}
          </p>
        </div>

        <div className="detail-hero-metrics">
          <div>
            <p className="muted-label">Последнее закрытие</p>
            <strong className="detail-price">
              {formatPrice(latestPrice?.close_price, data.asset.currency)}
            </strong>
          </div>
          <div className="tag-row">
            <span className={`badge badge-${dayChangeVariant}`}>
              {dayChange === null ? "-" : `${formatPercent(dayChange)} day`}
            </span>
            {latestIndicator ? (
              <span className="badge badge-muted">{formatPercent(latestIndicator.weekly_return)} week</span>
            ) : null}
          </div>
        </div>
      </section>

      <div className="detail-layout">
        <div className="detail-main">
          <PricesBlock prices={data.prices} ticker={data.asset.ticker} />
          <IndicatorsBlock indicators={data.indicators} ticker={data.asset.ticker} />

          {latestForecast ? (
            <ForecastsBlock forecasts={[latestForecast]} title="Latest Forecast" />
          ) : (
            <EmptyState
              description="Backend пока не вернул свежий прогноз по этому инструменту в dashboard summary."
              title="Прогноз для актива пока недоступен"
            />
          )}

          <EventsListBlock items={data.relatedEvents} kind="events" title="Related News & Events" />
        </div>

        <aside className="detail-side">
          <RegimeBlock compact regime={data.regime} />

          <section className="card">
            <div className="section-header">
              <h2 className="section-title">Key Stats</h2>
            </div>

            <div className="info-list">
              <div className="info-row">
                <span>Тикер</span>
                <strong>{data.asset.ticker}</strong>
              </div>
              <div className="info-row">
                <span>Тип</span>
                <strong>{data.asset.asset_type || "-"}</strong>
              </div>
              <div className="info-row">
                <span>Источник</span>
                <strong>Backend API</strong>
              </div>
              <div className="info-row">
                <span>Последнее обновление</span>
                <strong>{formatDateTime(latestIndicator?.indicator_time || latestPrice?.candle_time)}</strong>
              </div>
            </div>
          </section>

          <section className="card">
            <div className="section-header">
              <h2 className="section-title">Forecast History</h2>
            </div>

            {data.forecasts.length > 0 ? (
              <div className="stack-list">
                {data.forecasts.map((forecast) => (
                  <div className="list-item list-item-compact" key={forecast.id}>
                    <div className="list-item-head">
                      <strong>{formatDateTime(forecast.generated_at)}</strong>
                      <span className="badge badge-muted">{forecast.horizon}</span>
                    </div>
                    <p className="muted-text">{forecast.explanation}</p>
                  </div>
                ))}
              </div>
            ) : (
              <div className="inline-empty">История прогнозов по активу пока не сформирована.</div>
            )}
          </section>

          <section className="card">
            <div className="section-header">
              <h2 className="section-title">What to monitor</h2>
            </div>

            <div className="info-list">
              <div className="info-row">
                <span>RSI</span>
                <strong>{latestIndicator?.rsi ?? "-"}</strong>
              </div>
              <div className="info-row">
                <span>Trend</span>
                <strong>{getTrendLabel(latestIndicator?.trend_direction)}</strong>
              </div>
              <div className="info-row">
                <span>Связанных событий</span>
                <strong>{data.relatedEvents.length}</strong>
              </div>
              <div className="info-row">
                <span>Недельная доходность</span>
                <strong>{formatPercent(latestIndicator?.weekly_return)}</strong>
              </div>
            </div>
          </section>
        </aside>
      </div>
    </div>
  );
}
