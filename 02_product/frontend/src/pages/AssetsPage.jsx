import { useState } from "react";
import { api } from "../api/client";
import { AssetCard } from "../components/AssetCard";
import { EmptyState } from "../components/EmptyState";
import { LoadingBlock } from "../components/LoadingBlock";
import { NoticeCard } from "../components/NoticeCard";
import { PageHeader } from "../components/PageHeader";
import { getAssetTypeLabel } from "../lib/formatters";
import { useRemoteData } from "../lib/useRemoteData";

export function AssetsPage() {
  const [query, setQuery] = useState("");
  const [activeType, setActiveType] = useState("all");

  const { data, error, loading, reload } = useRemoteData(async () => {
    const assetsResponse = await api.getAssets();
    let dashboardSummary = null;
    let notice = "";

    try {
      dashboardSummary = await api.getDashboardSummary();
    } catch (loadError) {
      notice =
        "Список активов загружен, но обогащение карточек из dashboard summary временно недоступно.";
    }

    const indicatorsByTicker = new Map(
      (dashboardSummary?.assets || []).map((item) => [item.ticker, item.latest_indicator || null])
    );
    const forecastsByTicker = new Map(
      (dashboardSummary?.latest_forecasts || []).map((item) => [item.asset_ticker, item])
    );

    return {
      items: (assetsResponse.items || []).map((asset) => ({
        ...asset,
        latest_forecast: forecastsByTicker.get(asset.ticker) || null,
        latest_indicator: indicatorsByTicker.get(asset.ticker) || null
      })),
      notice
    };
  }, []);

  if (loading) {
    return <LoadingBlock label="Загрузка списка активов..." />;
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
        title="Не удалось загрузить активы"
        variant="error"
      />
    );
  }

  const items = data.items || [];
  const assetTypes = ["all", ...new Set(items.map((asset) => getAssetTypeLabel(asset.asset_type)))];
  const filteredAssets = items.filter((asset) => {
    const matchesType = activeType === "all" || getAssetTypeLabel(asset.asset_type) === activeType;
    const haystack = `${asset.ticker} ${asset.name} ${asset.sector}`.toLowerCase();
    const matchesQuery = !query.trim() || haystack.includes(query.trim().toLowerCase());
    return matchesType && matchesQuery;
  });

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Каталог"
        actions={
          <button className="secondary-button" onClick={reload} type="button">
            Обновить
          </button>
        }
        description="Каталог инструментов и внешних факторов первой версии платформы, оформленный по desktop references."
        title="Активы"
      />

      {data.notice ? <NoticeCard description={data.notice} title="Частичная деградация данных" /> : null}

      <section className="toolbar-card card">
        <label className="search-field">
          <span className="search-icon" aria-hidden="true">
            ⌕
          </span>
          <input
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Поиск по тикеру или названию"
            type="search"
            value={query}
          />
        </label>

        <div className="chip-group">
          {assetTypes.map((type) => (
            <button
              key={type}
              className={`filter-chip${activeType === type ? " filter-chip-active" : ""}`}
              onClick={() => setActiveType(type)}
              type="button"
            >
              {type === "all" ? "Все" : type}
            </button>
          ))}
        </div>
      </section>

      <section className="kpi-grid">
        <article className="kpi-card">
          <span className="kpi-label">Активы</span>
          <strong className="kpi-value">{filteredAssets.length}</strong>
          <span className="kpi-meta">отфильтрованных карточек</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Индекс</span>
          <strong className="kpi-value">
            {items.filter((asset) => getAssetTypeLabel(asset.asset_type) === "Индекс").length}
          </strong>
          <span className="kpi-meta">индексных инструментов</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Акции</span>
          <strong className="kpi-value">
            {items.filter((asset) => getAssetTypeLabel(asset.asset_type) === "Акции").length}
          </strong>
          <span className="kpi-meta">акций в первой версии</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Сырье</span>
          <strong className="kpi-value">
            {items.filter((asset) => getAssetTypeLabel(asset.asset_type) === "Сырье").length}
          </strong>
          <span className="kpi-meta">внешних сырьевых факторов</span>
        </article>
      </section>

      {filteredAssets.length === 0 ? (
        <EmptyState
          action={
            <button
              className="secondary-button"
              onClick={() => {
                setActiveType("all");
                setQuery("");
              }}
              type="button"
            >
              Сбросить фильтры
            </button>
          }
          description="Попробуйте снять фильтры или изменить поисковый запрос."
          title="По текущим условиям активы не найдены"
        />
      ) : null}

      <div className="cards-grid">
        {filteredAssets.map((asset) => (
          <AssetCard asset={asset} key={asset.id} />
        ))}
      </div>
    </div>
  );
}
