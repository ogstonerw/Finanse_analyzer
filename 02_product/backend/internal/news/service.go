package news

import (
	"context"
	"errors"
	"time"

	"diploma-market-ai/02_product/backend/internal/collectors"
	"diploma-market-ai/02_product/backend/internal/storage"
)

var ErrNewsItemNotFound = errors.New("news item not found")

type Service struct {
	sourcesRepository *storage.SourcesRepository
	newsRepository    *storage.NewsItemsRepository
	collector         collectors.NewsCollector
}

type Item struct {
	ID          string    `json:"id"`
	SourceID    string    `json:"source_id"`
	SourceName  string    `json:"source_name"`
	ExternalID  string    `json:"external_id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary,omitempty"`
	Body        string    `json:"body,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	CollectedAt time.Time `json:"collected_at"`
	URL         string    `json:"url"`
}

func NewService(store *storage.Postgres, collector collectors.NewsCollector) *Service {
	return &Service{
		sourcesRepository: storage.NewSourcesRepository(store),
		newsRepository:    storage.NewNewsItemsRepository(store),
		collector:         collector,
	}
}

func (s *Service) SyncLatest(ctx context.Context) error {
	if s.collector == nil {
		return errors.New("news collector is not configured")
	}

	source, err := s.sourcesRepository.GetByBaseURL(ctx, s.collector.SourceBaseURL())
	if err != nil {
		return err
	}

	collected, err := s.collector.CollectLatest(ctx)
	if err != nil {
		return err
	}

	items := make([]storage.UpsertNewsItemParams, 0, len(collected))
	for _, item := range collected {
		items = append(items, storage.UpsertNewsItemParams{
			SourceID:    source.ID,
			ExternalID:  item.ExternalID,
			Title:       item.Title,
			Summary:     item.Summary,
			Body:        item.Body,
			PublishedAt: item.PublishedAt,
			CollectedAt: item.CollectedAt,
			URL:         item.URL,
		})
	}

	return s.newsRepository.UpsertBatch(ctx, items)
}

func (s *Service) List(ctx context.Context) ([]Item, error) {
	items, err := s.newsRepository.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Item, 0, len(items))
	for _, item := range items {
		result = append(result, Item{
			ID:          item.ID,
			SourceID:    item.SourceID,
			SourceName:  item.SourceName,
			ExternalID:  item.ExternalID,
			Title:       item.Title,
			Summary:     item.Summary,
			PublishedAt: item.PublishedAt,
			CollectedAt: item.CollectedAt,
			URL:         item.URL,
		})
	}

	return result, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Item, error) {
	item, err := s.newsRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNewsItemNotFound) {
			return Item{}, ErrNewsItemNotFound
		}
		return Item{}, err
	}

	return Item{
		ID:          item.ID,
		SourceID:    item.SourceID,
		SourceName:  item.SourceName,
		ExternalID:  item.ExternalID,
		Title:       item.Title,
		Summary:     item.Summary,
		Body:        item.Body,
		PublishedAt: item.PublishedAt,
		CollectedAt: item.CollectedAt,
		URL:         item.URL,
	}, nil
}
