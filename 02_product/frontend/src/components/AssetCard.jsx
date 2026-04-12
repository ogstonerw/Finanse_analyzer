import { Link } from "react-router-dom";
import {
  formatPercent,
  formatPrice,
  getAssetTypeLabel,
  getConfidenceLabel,
  getDirectionClassName,
  getDirectionLabel,
  getTrendLabel
} from "../lib/formatters";

export function AssetCard({ asset }) {
  const weeklyReturn = asset.latest_indicator?.weekly_return;
  const trend = asset.latest_indicator?.trend_direction;
  const forecast = asset.latest_forecast || null;
  const priceValue =
    asset.last_price ??
    asset.latest_price ??
    asset.close_price ??
    asset.last_close_price ??
    null;
  const priceLabel = formatPrice(priceValue, asset.currency);
  const changeVariant =
    weeklyReturn === null || weeklyReturn === undefined ? "muted" : weeklyReturn >= 0 ? "positive" : "negative";

  return (
    <article className="card asset-card">
      <div className="asset-card-head">
        <div>
          <p className="asset-ticker">{asset.ticker}</p>
          <h3 className="section-title">{asset.name}</h3>
        </div>
        <span className="badge badge-muted">{getAssetTypeLabel(asset.asset_type)}</span>
      </div>

      <div className="asset-price-row">
        <div>
          <p className="muted-label">Последнее значение</p>
          <strong className="asset-price-value">{priceLabel}</strong>
        </div>
        <span className={`value-pill value-pill-${changeVariant}`}>
          {formatPercent(weeklyReturn)}
        </span>
      </div>

      <dl className="inline-metrics">
        <div>
          <dt>Сектор</dt>
          <dd>{asset.sector || "-"}</dd>
        </div>
        <div>
          <dt>RSI</dt>
          <dd>{asset.latest_indicator?.rsi ?? "-"}</dd>
        </div>
        <div>
          <dt>Волатильность</dt>
          <dd>{asset.latest_indicator?.volatility ?? "-"}</dd>
        </div>
        <div>
          <dt>Тренд</dt>
          <dd>{getTrendLabel(trend)}</dd>
        </div>
      </dl>

      <div className="tag-row">
        {forecast ? (
          <>
            <span className={`badge badge-${getDirectionClassName(forecast.direction)}`}>
              {getDirectionLabel(forecast.direction)}
            </span>
            <span className="badge badge-muted">
              Уверенность: {getConfidenceLabel(forecast.confidence)}
            </span>
          </>
        ) : (
          <span className="badge badge-muted">Прогноз пока не готов</span>
        )}
      </div>

      <Link className="text-link" to={`/assets/${asset.ticker}`}>
        Открыть карточку
      </Link>
    </article>
  );
}
