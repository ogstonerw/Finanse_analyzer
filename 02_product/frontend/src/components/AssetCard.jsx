import { Link } from "react-router-dom";
import { formatPercent, getTrendLabel } from "../lib/formatters";

export function AssetCard({ asset }) {
  const weeklyReturn = asset.latest_indicator?.weekly_return;
  const trend = asset.latest_indicator?.trend_direction;

  return (
    <article className="card asset-card">
      <div className="asset-card-head">
        <div>
          <p className="asset-ticker">{asset.ticker}</p>
          <h3 className="section-title">{asset.name}</h3>
        </div>
        <span className="badge badge-muted">{asset.asset_type}</span>
      </div>

      <dl className="inline-metrics">
        <div>
          <dt>Сектор</dt>
          <dd>{asset.sector}</dd>
        </div>
        <div>
          <dt>Валюта</dt>
          <dd>{asset.currency}</dd>
        </div>
        <div>
          <dt>Недельная доходность</dt>
          <dd>{formatPercent(weeklyReturn)}</dd>
        </div>
        <div>
          <dt>Тренд</dt>
          <dd>{getTrendLabel(trend)}</dd>
        </div>
      </dl>

      <Link className="text-link" to={`/assets/${asset.ticker}`}>
        Открыть детали актива
      </Link>
    </article>
  );
}
