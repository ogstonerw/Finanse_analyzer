import { formatDateTime, getEventTypeLabel } from "../lib/formatters";
import { EmptyState } from "./EmptyState";

export function EventsListBlock({ items, kind, title }) {
  if (!items?.length) {
    return (
      <section className="card">
        <div className="section-header">
          <h2 className="section-title">{title}</h2>
          <span className="badge badge-muted">0</span>
        </div>

        <EmptyState
          description="Backend пока не вернул записей для этого блока."
          title="Нет данных для отображения"
        />
      </section>
    );
  }

  return (
    <section className="card">
      <div className="section-header">
        <h2 className="section-title">{title}</h2>
        <span className="badge badge-muted">{items.length}</span>
      </div>

      <div className="stack-list">
        {items.map((item) => {
          const itemKind = kind === "mixed" ? item.kind : kind;
          const typeLabel = itemKind === "news" ? "Новость" : getEventTypeLabel(item.event_type);
          const titleText = itemKind === "news" ? item.title : item.news_title || item.title;
          const summaryText =
            itemKind === "news"
              ? item.summary || item.body || "Нет краткого описания."
              : item.summary || "Backend не вернул расширенное описание события.";

          return (
            <article className="list-item event-card" key={`${itemKind}-${item.id}`}>
              <div className="list-item-head">
                <div>
                  <p className="muted-label">{formatDateTime(item.published_at)}</p>
                  <h3>{titleText}</h3>
                </div>
                <span className={`badge badge-${itemKind === "news" ? "info" : "muted"}`}>
                  {typeLabel}
                </span>
              </div>

              <p>{summaryText}</p>

              <div className="tag-row">
                {item.asset_ticker ? <span className="badge badge-muted">{item.asset_ticker}</span> : null}
                {item.asset_name ? <span className="badge badge-muted">{item.asset_name}</span> : null}
                {item.source_name ? <span className="badge badge-muted">{item.source_name}</span> : null}
              </div>
            </article>
          );
        })}
      </div>
    </section>
  );
}
