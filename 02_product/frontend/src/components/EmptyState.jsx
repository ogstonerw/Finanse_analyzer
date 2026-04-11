export function EmptyState({ title, description }) {
  return (
    <div className="card empty-state">
      <h3 className="section-title">{title}</h3>
      <p>{description}</p>
    </div>
  );
}
