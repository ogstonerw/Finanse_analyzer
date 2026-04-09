package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrNewsItemNotFound = errors.New("news item not found")

type NewsItemRecord struct {
	ID          string
	SourceID    string
	SourceName  string
	ExternalID  string
	Title       string
	Summary     string
	Body        string
	PublishedAt time.Time
	CollectedAt time.Time
	URL         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UpsertNewsItemParams struct {
	SourceID    string
	ExternalID  string
	Title       string
	Summary     string
	Body        string
	PublishedAt time.Time
	CollectedAt time.Time
	URL         string
}

type NewsItemsRepository struct {
	db *sql.DB
}

func NewNewsItemsRepository(store *Postgres) *NewsItemsRepository {
	return &NewsItemsRepository{db: store.DB()}
}

func (r *NewsItemsRepository) UpsertBatch(ctx context.Context, items []UpsertNewsItemParams) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin upsert news items tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const query = `
		INSERT INTO news_items (
			source_id, external_id, title, summary, body, published_at, collected_at, url
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (source_id, external_id)
		DO UPDATE SET
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			body = EXCLUDED.body,
			published_at = EXCLUDED.published_at,
			collected_at = EXCLUDED.collected_at,
			url = EXCLUDED.url,
			updated_at = NOW()
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare upsert news items: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err = stmt.ExecContext(
			ctx,
			item.SourceID,
			item.ExternalID,
			item.Title,
			item.Summary,
			item.Body,
			item.PublishedAt,
			item.CollectedAt,
			item.URL,
		); err != nil {
			return fmt.Errorf("exec upsert news item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit upsert news items: %w", err)
	}

	return nil
}

func (r *NewsItemsRepository) List(ctx context.Context) ([]NewsItemRecord, error) {
	const query = `
		SELECT
			ni.id,
			ni.source_id,
			s.name,
			ni.external_id,
			ni.title,
			COALESCE(ni.summary, ''),
			COALESCE(ni.body, ''),
			ni.published_at,
			ni.collected_at,
			ni.url,
			ni.created_at,
			ni.updated_at
		FROM news_items ni
		INNER JOIN sources s ON s.id = ni.source_id
		ORDER BY ni.published_at DESC, ni.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list news items: %w", err)
	}
	defer rows.Close()

	items := make([]NewsItemRecord, 0)
	for rows.Next() {
		var item NewsItemRecord
		if err := rows.Scan(
			&item.ID,
			&item.SourceID,
			&item.SourceName,
			&item.ExternalID,
			&item.Title,
			&item.Summary,
			&item.Body,
			&item.PublishedAt,
			&item.CollectedAt,
			&item.URL,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan news item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate news items: %w", err)
	}

	return items, nil
}

func (r *NewsItemsRepository) GetByID(ctx context.Context, id string) (NewsItemRecord, error) {
	const query = `
		SELECT
			ni.id,
			ni.source_id,
			s.name,
			ni.external_id,
			ni.title,
			COALESCE(ni.summary, ''),
			COALESCE(ni.body, ''),
			ni.published_at,
			ni.collected_at,
			ni.url,
			ni.created_at,
			ni.updated_at
		FROM news_items ni
		INNER JOIN sources s ON s.id = ni.source_id
		WHERE ni.id = $1
	`

	var item NewsItemRecord
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.SourceID,
		&item.SourceName,
		&item.ExternalID,
		&item.Title,
		&item.Summary,
		&item.Body,
		&item.PublishedAt,
		&item.CollectedAt,
		&item.URL,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewsItemRecord{}, ErrNewsItemNotFound
		}
		return NewsItemRecord{}, fmt.Errorf("get news item by id: %w", err)
	}

	return item, nil
}
