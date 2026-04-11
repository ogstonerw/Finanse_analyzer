import { api } from "../api/client";
import { AssetCard } from "../components/AssetCard";
import { DashboardSummaryBlock } from "../components/DashboardSummaryBlock";
import { EmptyState } from "../components/EmptyState";
import { EventsListBlock } from "../components/EventsListBlock";
import { ForecastsBlock } from "../components/ForecastsBlock";
import { LoadingBlock } from "../components/LoadingBlock";
import { PageHeader } from "../components/PageHeader";
import { RegimeBlock } from "../components/RegimeBlock";
import { useRemoteData } from "../lib/useRemoteData";

export function DashboardPage() {
  const { data, error, loading, reload } = useRemoteData(
    async () => {
      const [summary, regime] = await Promise.all([
        api.getDashboardSummary(),
        api.getCurrentRegime()
      ]);

      return {
        ...summary,
        regime: regime || summary.regime
      };
    },
    []
  );

  if (loading) {
    return <LoadingBlock label="Загрузка dashboard..." />;
  }

  if (error) {
    return <EmptyState description={error} title="Не удалось загрузить dashboard" />;
  }

  return (
    <div className="page-stack">
      <PageHeader
        actions={
          <button className="secondary-button" onClick={reload} type="button">
            Обновить сводку
          </button>
        }
        description="Сводный экран для предзащиты: общий режим рынка, активы, события и последние сигналы."
        title="Дашборд платформы"
      />

      <DashboardSummaryBlock summary={data} />
      <RegimeBlock regime={data.regime} />

      <section className="page-section">
        <div className="section-header">
          <h2 className="section-title">Активы в наблюдении</h2>
          <span className="badge badge-muted">{data.assets.length}</span>
        </div>

        <div className="cards-grid">
          {data.assets.map((asset) => (
            <AssetCard asset={asset} key={asset.id} />
          ))}
        </div>
      </section>

      <ForecastsBlock forecasts={data.latest_forecasts} title="Последние прогнозы" />
      <EventsListBlock items={data.recent_events} kind="events" title="Последние события" />
    </div>
  );
}
