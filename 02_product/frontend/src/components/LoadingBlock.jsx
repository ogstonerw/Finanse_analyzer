export function LoadingBlock({ label = "Загрузка данных..." }) {
  return (
    <div className="card loading-block">
      <div aria-hidden="true" className="loading-spinner" />
      <p>{label}</p>
    </div>
  );
}
