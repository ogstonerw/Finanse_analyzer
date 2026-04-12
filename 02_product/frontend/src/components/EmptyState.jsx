export function EmptyState({ action, description, title, variant = "neutral" }) {
  return (
    <div className={`card empty-state empty-state-${variant}`}>
      <div className={`empty-state-icon empty-state-icon-${variant}`} aria-hidden="true">
        ?
      </div>
      <h3 className="section-title">{title}</h3>
      <p className="muted-text">{description}</p>
      {action ? <div className="empty-state-action">{action}</div> : null}
    </div>
  );
}
