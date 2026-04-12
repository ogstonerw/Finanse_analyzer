import { formatDateTime, formatNumber, formatPrice } from "../lib/formatters";

function buildChartPoints(prices) {
  if (!prices.length) {
    return "";
  }

  const closes = prices.map((item) => Number(item.close_price ?? item.close ?? 0));
  const min = Math.min(...closes);
  const max = Math.max(...closes);
  const range = max - min || 1;

  return closes
    .map((value, index) => {
      const x = (index / Math.max(closes.length - 1, 1)) * 100;
      const y = 100 - ((value - min) / range) * 100;
      return `${x},${y}`;
    })
    .join(" ");
}

export function PricesBlock({ prices, ticker }) {
  const recent = prices?.slice(-10).reverse() || [];
  const chartSeries = prices?.slice(-24) || [];
  const latest = chartSeries[chartSeries.length - 1] || null;
  const previous = chartSeries[chartSeries.length - 2] || null;
  const points = buildChartPoints(chartSeries);
  const dayChange =
    latest && previous && Number(previous.close_price)
      ? ((Number(latest.close_price) - Number(previous.close_price)) / Number(previous.close_price)) * 100
      : null;
  const dayChangeVariant =
    dayChange === null || dayChange === undefined ? "muted" : dayChange >= 0 ? "positive" : "negative";

  return (
    <section className="card">
      <div className="section-header">
        <div>
          <h2 className="section-title">История свечей</h2>
          <p className="section-subtitle">
            {ticker} · последние {recent.length || 0} записей
          </p>
        </div>
      </div>

      {chartSeries.length > 0 ? (
        <>
          <div className="chart-card">
            <div className="chart-card-head">
              <div>
                <p className="muted-label">Последнее закрытие</p>
                <strong className="asset-price-value">{formatPrice(latest.close_price)}</strong>
              </div>
              <div className="tag-row">
                <span className={`badge badge-${dayChangeVariant}`}>
                  {dayChange === null ? "-" : `${formatNumber(dayChange)}% за день`}
                </span>
              </div>
            </div>

            <div className="chart-canvas">
              <svg aria-label={`График цены ${ticker}`} className="sparkline-chart" viewBox="0 0 100 100">
                <polyline points={points} />
              </svg>
            </div>

            <div className="chart-footnote">
              O {formatNumber(latest.open_price)} · H {formatNumber(latest.high_price)} · L{" "}
              {formatNumber(latest.low_price)} · C {formatNumber(latest.close_price)}
            </div>
          </div>

          <div className="table-wrapper">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Дата</th>
                  <th>Открытие</th>
                  <th>Максимум</th>
                  <th>Минимум</th>
                  <th>Закрытие</th>
                  <th>Объем</th>
                </tr>
              </thead>
              <tbody>
                {recent.map((item) => (
                  <tr key={item.candle_time}>
                    <td>{formatDateTime(item.candle_time)}</td>
                    <td>{formatNumber(item.open_price)}</td>
                    <td>{formatNumber(item.high_price)}</td>
                    <td>{formatNumber(item.low_price)}</td>
                    <td>{formatNumber(item.close_price)}</td>
                    <td>{formatNumber(item.volume)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      ) : (
        <div className="inline-empty">Свечи пока не найдены.</div>
      )}
    </section>
  );
}
