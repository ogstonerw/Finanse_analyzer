export function EmptyState({ action, description, title }) {
  return (
    <div className="card empty-state">
      <div className="empty-state-icon" aria-hidden="true">
        ?
      </div>
      <h3 className="section-title">{title}</h3>
      <p className="muted-text">{description}</p>
      {action ? <div className="empty-state-action">{action}</div> : null}
    </div>
  );
}
