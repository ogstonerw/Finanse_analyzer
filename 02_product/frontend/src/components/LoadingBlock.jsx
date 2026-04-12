export function LoadingBlock({ label = "Загрузка данных..." }) {
  return (
    <div className="card loading-block">
      <div className="loading-block-header">
        <div aria-hidden="true" className="loading-spinner" />
        <p>{label}</p>
      </div>

      <div aria-hidden="true" className="skeleton-stack">
        <span className="skeleton-line skeleton-line-lg" />
        <span className="skeleton-line skeleton-line-md" />
        <span className="skeleton-line skeleton-line-sm" />
      </div>
    </div>
  );
}
