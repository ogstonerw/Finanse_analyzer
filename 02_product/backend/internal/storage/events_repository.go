package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrEventNotFound = errors.New("event not found")

type EventRecord struct {
	ID          string
	NewsItemID  string
	NewsTitle   string
	AssetID     sql.NullString
	AssetTicker sql.NullString
	AssetName   sql.NullString
	EventType   string
	Summary     string
	ExtractedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UpsertEventParams struct {
	NewsItemID  string
	AssetID     *string
	EventType   string
	Summary     string
	ExtractedAt time.Time
}

type EventsRepository struct {
	db *sql.DB
}

func NewEventsRepository(store *Postgres) *EventsRepository {
	return &EventsRepository{db: store.DB()}
}

func (r *EventsRepository) UpsertBatch(ctx context.Context, items []UpsertEventParams) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin upsert events tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const query = `
		INSERT INTO events (
			news_item_id, asset_id, event_type, summary, extracted_at
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (news_item_id)
		DO UPDATE SET
			asset_id = EXCLUDED.asset_id,
			event_type = EXCLUDED.event_type,
			summary = EXCLUDED.summary,
			extracted_at = EXCLUDED.extracted_at,
			updated_at = NOW()
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare upsert events: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err = stmt.ExecContext(
			ctx,
			item.NewsItemID,
			item.AssetID,
			item.EventType,
			item.Summary,
			item.ExtractedAt,
		); err != nil {
			return fmt.Errorf("exec upsert event: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit upsert events: %w", err)
	}

	return nil
}

func (r *EventsRepository) List(ctx context.Context) ([]EventRecord, error) {
	const query = `
		SELECT
			e.id,
			e.news_item_id,
			ni.title,
			e.asset_id,
			a.ticker,
			a.name,
			e.event_type,
			e.summary,
			e.extracted_at,
			e.created_at,
			e.updated_at
		FROM events e
		INNER JOIN news_items ni ON ni.id = e.news_item_id
		LEFT JOIN assets a ON a.id = e.asset_id
		ORDER BY e.extracted_at DESC, e.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	items := make([]EventRecord, 0)
	for rows.Next() {
		var item EventRecord
		if err := rows.Scan(
			&item.ID,
			&item.NewsItemID,
			&item.NewsTitle,
			&item.AssetID,
			&item.AssetTicker,
			&item.AssetName,
			&item.EventType,
			&item.Summary,
			&item.ExtractedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return items, nil
}

func (r *EventsRepository) GetByID(ctx context.Context, id string) (EventRecord, error) {
	const query = `
		SELECT
			e.id,
			e.news_item_id,
			ni.title,
			e.asset_id,
			a.ticker,
			a.name,
			e.event_type,
			e.summary,
			e.extracted_at,
			e.created_at,
			e.updated_at
		FROM events e
		INNER JOIN news_items ni ON ni.id = e.news_item_id
		LEFT JOIN assets a ON a.id = e.asset_id
		WHERE e.id = $1
	`

	var item EventRecord
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.NewsItemID,
		&item.NewsTitle,
		&item.AssetID,
		&item.AssetTicker,
		&item.AssetName,
		&item.EventType,
		&item.Summary,
		&item.ExtractedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EventRecord{}, ErrEventNotFound
		}
		return EventRecord{}, fmt.Errorf("get event by id: %w", err)
	}

	return item, nil
}

func (r *EventsRepository) GetLatestRelevant(ctx context.Context, assetID string) (EventRecord, error) {
	const query = `
		SELECT
			e.id,
			e.news_item_id,
			ni.title,
			e.asset_id,
			a.ticker,
			a.name,
			e.event_type,
			e.summary,
			e.extracted_at,
			e.created_at,
			e.updated_at
		FROM events e
		INNER JOIN news_items ni ON ni.id = e.news_item_id
		LEFT JOIN assets a ON a.id = e.asset_id
		WHERE e.asset_id = $1 OR e.asset_id IS NULL
		ORDER BY
			CASE WHEN e.asset_id = $1 THEN 0 ELSE 1 END,
			e.extracted_at DESC,
			e.created_at DESC
		LIMIT 1
	`

	var item EventRecord
	err := r.db.QueryRowContext(ctx, query, assetID).Scan(
		&item.ID,
		&item.NewsItemID,
		&item.NewsTitle,
		&item.AssetID,
		&item.AssetTicker,
		&item.AssetName,
		&item.EventType,
		&item.Summary,
		&item.ExtractedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EventRecord{}, ErrEventNotFound
		}
		return EventRecord{}, fmt.Errorf("get latest relevant event: %w", err)
	}

	return item, nil
}
