export function NoticeCard({ description, title = "Замечание" }) {
  return (
    <section className="card notice-card">
      <div className="notice-card-icon" aria-hidden="true">
        !
      </div>
      <div className="notice-card-copy">
        <h2 className="section-title">{title}</h2>
        <p>{description}</p>
      </div>
    </section>
  );
}
