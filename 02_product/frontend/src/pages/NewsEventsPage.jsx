import { api } from "../api/client";
import { EmptyState } from "../components/EmptyState";
import { EventsListBlock } from "../components/EventsListBlock";
import { LoadingBlock } from "../components/LoadingBlock";
import { PageHeader } from "../components/PageHeader";
import { useRemoteData } from "../lib/useRemoteData";

export function NewsEventsPage() {
  const { data, loading, error, reload } = useRemoteData(
    async () => {
      const [news, events] = await Promise.all([api.getNews(), api.getEvents()]);

      return {
        events: events.items || [],
        news: news.items || []
      };
    },
    []
  );

  if (loading) {
    return <LoadingBlock label="Загрузка новостей и событий..." />;
  }

  if (error) {
    return <EmptyState title="Не удалось загрузить новости и события" description={error} />;
  }

  return (
    <div className="page-stack">
      <PageHeader
        title="Новости и события"
        description="Раздельный просмотр нормализованных новостей и событий, уже подготовленных backend для MVP-демонстрации."
        actions={
          <button className="secondary-button" onClick={reload} type="button">
            Обновить
          </button>
        }
      />

      <div className="two-column-grid">
        <EventsListBlock items={data.news} kind="news" title="Новости" />
        <EventsListBlock items={data.events} kind="events" title="События" />
      </div>
    </div>
  );
}
