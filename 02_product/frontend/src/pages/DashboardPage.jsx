import { Link } from "react-router-dom";
import { api } from "../api/client";
import { AssetCard } from "../components/AssetCard";
import { DashboardSummaryBlock } from "../components/DashboardSummaryBlock";
import { EmptyState } from "../components/EmptyState";
import { EventsListBlock } from "../components/EventsListBlock";
import { ForecastsBlock } from "../components/ForecastsBlock";
import { LoadingBlock } from "../components/LoadingBlock";
import { NoticeCard } from "../components/NoticeCard";
import { PageHeader } from "../components/PageHeader";
import { RegimeBlock } from "../components/RegimeBlock";
import { useRemoteData } from "../lib/useRemoteData";

export function DashboardPage() {
  const { data, error, loading, reload } = useRemoteData(
    async () => {
      const summary = await api.getDashboardSummary();
      let regime = summary.regime || null;
      let notice = "";

      try {
        const nextRegime = await api.getCurrentRegime();
        regime = nextRegime || regime;
      } catch (loadError) {
        if (regime) {
          notice =
            "Отдельный endpoint режима рынка временно недоступен. Для демонстрации показано значение из dashboard summary.";
        } else {
          notice =
            "Backend не вернул актуальный режим рынка ни через dashboard summary, ни через отдельный endpoint.";
        }
      }

      return {
        ...summary,
        notice,
        regime
      };
    },
    []
  );

  if (loading) {
    return <LoadingBlock label="Загрузка dashboard..." />;
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
        title="Не удалось загрузить панель обзора"
        variant="error"
      />
    );
  }

  const forecasts = data.latest_forecasts || [];
  const assets = data.assets || [];
  const events = data.recent_events || [];

  const directionCounts = forecasts.reduce(
    (accumulator, forecast) => {
      accumulator[forecast.direction] = (accumulator[forecast.direction] || 0) + 1;
      return accumulator;
    },
    { negative: 0, neutral: 0, positive: 0 }
  );

  const strongestForecast = [...forecasts].sort(
    (left, right) =>
      (right.confidence || 0) * (right.strength || 0) - (left.confidence || 0) * (left.strength || 0)
  )[0];

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Обзор"
        actions={
          <button className="secondary-button" onClick={reload} type="button">
            Обновить сводку
          </button>
        }
        description="Desktop-first сводка для предзащиты: режим рынка, активы, последние сигналы и лента релевантных событий."
        title="Панель обзора"
      />

      {data.notice ? <NoticeCard description={data.notice} title="Частичная деградация данных" /> : null}

      <div className="dashboard-top-grid">
        <RegimeBlock regime={data.regime} />

        <div className="dashboard-side-stack">
          <DashboardSummaryBlock summary={data} />

          <section className="card">
            <div className="section-header">
              <h2 className="section-title">Быстрые действия</h2>
            </div>

            <div className="quick-actions-grid">
              <Link className="quick-action-card" to="/assets">
                <strong>Открыть активы</strong>
                <span>Каталог инструментов и внешних факторов первой версии.</span>
              </Link>
              <Link className="quick-action-card" to="/news">
                <strong>Открыть ленту</strong>
                <span>Лента новостей и событий с привязкой к наблюдаемым активам.</span>
              </Link>
              <Link className="quick-action-card" to="/forecasts">
                <strong>Посмотреть прогнозы</strong>
                <span>Последние сигналы, уверенность и объяснение модели.</span>
              </Link>
              <button className="quick-action-card quick-action-card-button" onClick={reload} type="button">
                <strong>Обновить dashboard</strong>
                <span>Повторный запрос к существующим backend endpoint&apos;ам.</span>
              </button>
            </div>
          </section>
        </div>
      </div>

      <section className="kpi-grid">
        <article className="kpi-card">
          <span className="kpi-label">Прогнозы</span>
          <strong className="kpi-value">{forecasts.length}</strong>
          <span className="kpi-meta">свежих сигналов в dashboard summary</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Активы</span>
          <strong className="kpi-value">{assets.length}</strong>
          <span className="kpi-meta">инструментов в наблюдении</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">События</span>
          <strong className="kpi-value">{events.length}</strong>
          <span className="kpi-meta">событий в последней выборке</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Статус</span>
          <strong className="kpi-value">{data.generated_at ? "OK" : "-"}</strong>
          <span className="kpi-meta">
            {data.generated_at ? "данные готовы к демонстрации" : "нет отметки времени"}
          </span>
        </article>
      </section>

      <section className="page-section">
        <div className="section-header">
          <div>
            <h2 className="section-title">Активы в фокусе</h2>
            <p className="section-subtitle">Ключевые инструменты, попавшие в dashboard summary.</p>
          </div>
          <Link className="text-link" to="/assets">
            Все активы
          </Link>
        </div>

        {assets.length > 0 ? (
          <div className="cards-grid">
            {assets.slice(0, 6).map((asset) => (
              <AssetCard asset={asset} key={asset.id} />
            ))}
          </div>
        ) : (
          <EmptyState
            description="Dashboard summary пока не вернул карточки активов для обзорного экрана."
            title="Активы в фокусе пока недоступны"
          />
        )}
      </section>

      <div className="two-column-grid two-column-grid-wide">
        <ForecastsBlock forecasts={forecasts.slice(0, 4)} title="Последние прогнозы" />
        <EventsListBlock items={events.slice(0, 4)} kind="events" title="Лента новостей и событий" />
      </div>

      <section className="kpi-grid">
        <article className="kpi-card">
          <span className="kpi-label">Позитивные сигналы</span>
          <strong className="kpi-value">{directionCounts.positive}</strong>
          <span className="kpi-meta">активов с ростом</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Негативные сигналы</span>
          <strong className="kpi-value">{directionCounts.negative}</strong>
          <span className="kpi-meta">активов со снижением</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Нейтральные сигналы</span>
          <strong className="kpi-value">{directionCounts.neutral}</strong>
          <span className="kpi-meta">активов без выраженного направления</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Сильнейший сигнал</span>
          <strong className="kpi-value">{strongestForecast?.asset_ticker || "-"}</strong>
          <span className="kpi-meta">{strongestForecast?.asset_name || "нет данных"}</span>
        </article>
      </section>
    </div>
  );
}
