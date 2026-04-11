import { formatDateTime, formatNumber, formatPercent, getTrendLabel } from "../lib/formatters";

function Metric({ title, value }) {
  return (
    <div className="metric-card">
      <span>{title}</span>
      <strong>{value}</strong>
    </div>
  );
}

export function IndicatorsBlock({ indicators, ticker }) {
  const latest = indicators?.[indicators.length - 1];
  const recent = indicators?.slice(-5).reverse() || [];

  return (
    <section className="card">
      <div className="section-header">
        <div>
          <h2 className="section-title">Технические индикаторы</h2>
          <p className="section-subtitle">{ticker}</p>
        </div>
      </div>

      {latest ? (
        <div className="metrics-grid">
          <Metric title="Последний расчет" value={formatDateTime(latest.indicator_time)} />
          <Metric title="Недельная доходность" value={formatPercent(latest.weekly_return)} />
          <Metric title="RSI" value={formatNumber(latest.rsi)} />
          <Metric title="Волатильность" value={formatNumber(latest.volatility)} />
          <Metric title="Тренд" value={getTrendLabel(latest.trend_direction)} />
          <Metric title="Позиция в канале" value={formatNumber(latest.channel_position)} />
        </div>
      ) : (
        <p>Индикаторы пока не найдены.</p>
      )}

      {recent.length > 0 ? (
        <div className="table-wrapper">
          <table className="data-table">
            <thead>
              <tr>
                <th>Дата</th>
                <th>Доходность</th>
                <th>RSI</th>
                <th>Волатильность</th>
                <th>Тренд</th>
                <th>Статус</th>
              </tr>
            </thead>
            <tbody>
              {recent.map((item) => (
                <tr key={item.indicator_time}>
                  <td>{formatDateTime(item.indicator_time)}</td>
                  <td>{formatPercent(item.weekly_return)}</td>
                  <td>{formatNumber(item.rsi)}</td>
                  <td>{formatNumber(item.volatility)}</td>
                  <td>{getTrendLabel(item.trend_direction)}</td>
                  <td>{item.calculation_status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </section>
  );
}
