import { useState } from "react";
import { api } from "../api/client";
import { EmptyState } from "../components/EmptyState";
import { EventsListBlock } from "../components/EventsListBlock";
import { LoadingBlock } from "../components/LoadingBlock";
import { NoticeCard } from "../components/NoticeCard";
import { PageHeader } from "../components/PageHeader";
import { getEventTypeLabel } from "../lib/formatters";
import { useRemoteData } from "../lib/useRemoteData";

export function NewsEventsPage() {
  const [query, setQuery] = useState("");
  const [activeKind, setActiveKind] = useState("all");

  const { data, loading, error, reload } = useRemoteData(
    async () => {
      const [newsResult, eventsResult] = await Promise.allSettled([api.getNews(), api.getEvents()]);
      const noticeParts = [];

      if (newsResult.status === "rejected" && eventsResult.status === "rejected") {
        throw new Error("Backend не вернул ни новости, ни события.");
      }

      if (newsResult.status === "rejected") {
        noticeParts.push("Новостная лента временно недоступна.");
      }

      if (eventsResult.status === "rejected") {
        noticeParts.push("Сервис структурированных событий временно недоступен.");
      }

      return {
        events: eventsResult.status === "fulfilled" ? eventsResult.value.items || [] : [],
        news: newsResult.status === "fulfilled" ? newsResult.value.items || [] : [],
        notice: noticeParts.join(" ")
      };
    },
    []
  );

  if (loading) {
    return <LoadingBlock label="Загрузка новостей и событий..." />;
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
        title="Не удалось загрузить новости и события"
        variant="error"
      />
    );
  }

  const combinedFeed = [
    ...(data.news || []).map((item) => ({
      ...item,
      kind: "news"
    })),
    ...(data.events || []).map((item) => ({
      ...item,
      kind: "events"
    }))
  ].sort((left, right) => new Date(right.published_at || 0) - new Date(left.published_at || 0));

  const eventTypes = ["all", ...new Set((data.events || []).map((item) => getEventTypeLabel(item.event_type)))];
  const filteredFeed = combinedFeed.filter((item) => {
    const typeLabel = item.kind === "events" ? getEventTypeLabel(item.event_type) : "News";
    const matchesKind =
      activeKind === "all" || activeKind === item.kind || (item.kind === "events" && activeKind === typeLabel);
    const title = item.kind === "news" ? item.title : item.news_title || item.title;
    const summary = item.summary || item.body || "";
    const haystack = `${title} ${summary} ${item.asset_ticker || ""} ${item.asset_name || ""}`.toLowerCase();
    const matchesQuery = !query.trim() || haystack.includes(query.trim().toLowerCase());
    return matchesKind && matchesQuery;
  });

  const assetMentions = filteredFeed.reduce((accumulator, item) => {
    if (item.asset_ticker) {
      accumulator[item.asset_ticker] = (accumulator[item.asset_ticker] || 0) + 1;
    }
    return accumulator;
  }, {});
  const topAsset = Object.entries(assetMentions).sort((left, right) => right[1] - left[1])[0];
  const activeKindLabel =
    activeKind === "all"
      ? "все записи"
      : activeKind === "news"
        ? "новости"
        : activeKind === "events"
          ? "события"
          : activeKind;

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Лента"
        title="Новости и события"
        description="Единая аналитическая лента с локальным поиском по уже нормализованным данным backend."
        actions={
          <button className="secondary-button" onClick={reload} type="button">
            Обновить
          </button>
        }
      />

      {data.notice ? <NoticeCard description={data.notice} title="Частичная деградация данных" /> : null}

      <section className="kpi-grid">
        <article className="kpi-card">
          <span className="kpi-label">Новости</span>
          <strong className="kpi-value">{data.news.length}</strong>
          <span className="kpi-meta">новостных записей</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">События</span>
          <strong className="kpi-value">{data.events.length}</strong>
          <span className="kpi-meta">структурированных событий</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Топ-актив</span>
          <strong className="kpi-value">{topAsset?.[0] || "-"}</strong>
          <span className="kpi-meta">{topAsset ? `${topAsset[1]} упоминаний` : "нет привязки к активам"}</span>
        </article>
        <article className="kpi-card">
          <span className="kpi-label">Типы событий</span>
          <strong className="kpi-value">{eventTypes.length - 1}</strong>
          <span className="kpi-meta">ключевых типов событий</span>
        </article>
      </section>

      <section className="toolbar-card card">
        <label className="search-field">
          <span className="search-icon" aria-hidden="true">
            ⌕
          </span>
          <input
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Поиск по заголовку или активу"
            type="search"
            value={query}
          />
        </label>

        <div className="chip-group">
          <button
            className={`filter-chip${activeKind === "all" ? " filter-chip-active" : ""}`}
            onClick={() => setActiveKind("all")}
            type="button"
          >
            Все
          </button>
          <button
            className={`filter-chip${activeKind === "news" ? " filter-chip-active" : ""}`}
            onClick={() => setActiveKind("news")}
            type="button"
          >
            Новости
          </button>
          <button
            className={`filter-chip${activeKind === "events" ? " filter-chip-active" : ""}`}
            onClick={() => setActiveKind("events")}
            type="button"
          >
            События
          </button>
          {eventTypes
            .filter((item) => item !== "all")
            .slice(0, 4)
            .map((type) => (
              <button
                key={type}
                className={`filter-chip${activeKind === type ? " filter-chip-active" : ""}`}
                onClick={() => setActiveKind(type)}
                type="button"
              >
                {type}
              </button>
            ))}
        </div>
      </section>

      <div className="news-layout">
        {filteredFeed.length > 0 ? (
          <EventsListBlock items={filteredFeed} kind="mixed" title="Общая лента" />
        ) : (
          <EmptyState
            action={
              <button
                className="secondary-button"
                onClick={() => {
                  setActiveKind("all");
                  setQuery("");
                }}
                type="button"
              >
                Сбросить фильтры
              </button>
            }
            description="Попробуйте снять фильтр или изменить поисковый запрос."
            title="По текущим условиям лента пуста"
          />
        )}

        <aside className="detail-side">
          <section className="card">
            <div className="section-header">
              <h2 className="section-title">Сводка ленты</h2>
            </div>

            <div className="info-list">
              <div className="info-row">
                <span>Записей в ленте</span>
                <strong>{filteredFeed.length}</strong>
              </div>
              <div className="info-row">
                <span>Новостей</span>
                <strong>{filteredFeed.filter((item) => item.kind === "news").length}</strong>
              </div>
              <div className="info-row">
                <span>Событий</span>
                <strong>{filteredFeed.filter((item) => item.kind === "events").length}</strong>
              </div>
              <div className="info-row">
                <span>Топ-актив</span>
                <strong>{topAsset?.[0] || "-"}</strong>
              </div>
            </div>
          </section>

          <section className="card">
            <div className="section-header">
              <h2 className="section-title">Активы в фокусе</h2>
            </div>

            <div className="stack-list">
              {Object.entries(assetMentions)
                .sort((left, right) => right[1] - left[1])
                .slice(0, 4)
                .map(([asset, count]) => (
                  <div className="list-item list-item-compact" key={asset}>
                    <div className="list-item-head">
                      <strong>{asset}</strong>
                      <span className="badge badge-muted">{count} записей</span>
                    </div>
                  </div>
                ))}

              {Object.keys(assetMentions).length === 0 ? (
                <div className="inline-empty">В текущей выборке нет связанных активов.</div>
              ) : null}
            </div>
          </section>

          <section className="card">
            <div className="section-header">
              <h2 className="section-title">Параметры фильтра</h2>
            </div>

            <div className="info-list">
              <div className="info-row">
                <span>Тип</span>
                <strong>{activeKindLabel}</strong>
              </div>
              <div className="info-row">
                <span>Запрос</span>
                <strong>{query || "без фильтра"}</strong>
              </div>
              <div className="info-row">
                <span>Источник</span>
                <strong>backend API</strong>
              </div>
            </div>
          </section>
        </aside>
      </div>
    </div>
  );
}
