import { formatDateTime } from "../lib/formatters";

export function EventsListBlock({ items, kind, title }) {
  return (
    <section className="card">
      <div className="section-header">
        <h2 className="section-title">{title}</h2>
        <span className="badge badge-muted">{items.length}</span>
      </div>

      <div className="stack-list">
        {items.map((item) => (
          <article className="list-item" key={item.id}>
            <div className="list-item-head">
              <h3>{kind === "news" ? item.title : item.news_title}</h3>
              <span className="badge">{kind === "news" ? "news" : item.event_type}</span>
            </div>

            <p className="muted-text">
              {formatDateTime(kind === "news" ? item.published_at : item.published_at)}
            </p>

            <p>{kind === "news" ? item.summary || item.body || "Без краткого описания." : item.summary}</p>

            {kind === "events" && item.asset_ticker ? (
              <p className="muted-text">
                Актив: {item.asset_ticker}
                {item.asset_name ? ` · ${item.asset_name}` : ""}
              </p>
            ) : null}
          </article>
        ))}
      </div>
    </section>
  );
}
