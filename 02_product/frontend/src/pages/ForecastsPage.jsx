import { api } from "../api/client";
import { EmptyState } from "../components/EmptyState";
import { ForecastsBlock } from "../components/ForecastsBlock";
import { LoadingBlock } from "../components/LoadingBlock";
import { PageHeader } from "../components/PageHeader";
import { RegimeBlock } from "../components/RegimeBlock";
import { useRemoteData } from "../lib/useRemoteData";

export function ForecastsPage() {
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

  const topForecasts = data.latestSingleForecast ? [data.latestSingleForecast] : [];

  return (
    <div className="page-stack">
      <PageHeader
        title="Последние прогнозы"
        description="Экран для демонстрации актуальных сигналов, их силы, уверенности и связи с текущим режимом рынка."
        actions={
          <button className="secondary-button" onClick={reload} type="button">
            Обновить
          </button>
        }
      />

      <RegimeBlock regime={data.regime} />

      <ForecastsBlock forecasts={topForecasts} title="Самый свежий прогноз" />
      <ForecastsBlock
        forecasts={data.latestForecasts}
        title="Последние прогнозы из dashboard summary"
      />
    </div>
  );
}
