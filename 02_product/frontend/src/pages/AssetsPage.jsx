import { api } from "../api/client";
import { AssetCard } from "../components/AssetCard";
import { EmptyState } from "../components/EmptyState";
import { LoadingBlock } from "../components/LoadingBlock";
import { PageHeader } from "../components/PageHeader";
import { useRemoteData } from "../lib/useRemoteData";

export function AssetsPage() {
  const { data, error, loading, reload } = useRemoteData(async () => {
    const [assetsResponse, dashboardSummary] = await Promise.all([
      api.getAssets(),
      api.getDashboardSummary()
    ]);

    const indicatorsByTicker = new Map(
      (dashboardSummary.assets || []).map((item) => [item.ticker, item.latest_indicator || null])
    );

    return (assetsResponse.items || []).map((asset) => ({
      ...asset,
      latest_indicator: indicatorsByTicker.get(asset.ticker) || null
    }));
  }, []);

  if (loading) {
    return <LoadingBlock label="Загрузка списка активов..." />;
  }

  if (error) {
    return <EmptyState description={error} title="Не удалось загрузить активы" />;
  }

  return (
    <div className="page-stack">
      <PageHeader
        actions={
          <button className="secondary-button" onClick={reload} type="button">
            Обновить
          </button>
        }
        description="Базовый каталог инструментов и внешних факторов первой версии платформы."
        title="Список активов"
      />

      <div className="cards-grid">
        {data.map((asset) => (
          <AssetCard asset={asset} key={asset.id} />
        ))}
      </div>
    </div>
  );
}
