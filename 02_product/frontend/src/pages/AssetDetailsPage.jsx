import { Link, useParams } from "react-router-dom";
import { api } from "../api/client";
import { EmptyState } from "../components/EmptyState";
import { ForecastsBlock } from "../components/ForecastsBlock";
import { IndicatorsBlock } from "../components/IndicatorsBlock";
import { LoadingBlock } from "../components/LoadingBlock";
import { PageHeader } from "../components/PageHeader";
import { PricesBlock } from "../components/PricesBlock";
import { RegimeBlock } from "../components/RegimeBlock";
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

  return (
    <div className="page-stack">
      <PageHeader
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
        description="Карточка актива с базовыми сведениями, свечами, индикаторами и рыночным режимом."
        title={`${data.asset.ticker} · ${data.asset.name}`}
      />

      <section className="card">
        <dl className="inline-metrics">
          <div>
            <dt>Тикер</dt>
            <dd>{data.asset.ticker}</dd>
          </div>
          <div>
            <dt>Тип</dt>
            <dd>{data.asset.asset_type}</dd>
          </div>
          <div>
            <dt>Сектор</dt>
            <dd>{data.asset.sector}</dd>
          </div>
          <div>
            <dt>Валюта</dt>
            <dd>{data.asset.currency}</dd>
          </div>
        </dl>
      </section>

      <RegimeBlock regime={data.regime} />
      <IndicatorsBlock indicators={data.indicators} ticker={data.asset.ticker} />
      <PricesBlock prices={data.prices} ticker={data.asset.ticker} />

      {data.forecasts.length > 0 ? (
        <ForecastsBlock forecasts={data.forecasts} title="Последние прогнозы по активу" />
      ) : (
        <EmptyState
          description="Backend пока не вернул свежие прогнозы по этому инструменту в dashboard summary."
          title="Прогнозов для актива пока нет"
        />
      )}
    </div>
  );
}
