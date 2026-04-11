import { formatDateTime, formatNumber } from "../lib/formatters";

export function PricesBlock({ prices, ticker }) {
  const recent = prices?.slice(-10).reverse() || [];

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

      {recent.length > 0 ? (
        <div className="table-wrapper">
          <table className="data-table">
            <thead>
              <tr>
                <th>Дата</th>
                <th>Open</th>
                <th>High</th>
                <th>Low</th>
                <th>Close</th>
                <th>Volume</th>
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
      ) : (
        <p>Свечи пока не найдены.</p>
      )}
    </section>
  );
}
