package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrSourceNotFound = errors.New("source not found")

type SourceRecord struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	BaseURL         string    `json:"base_url"`
	SourceType      string    `json:"source_type"`
	AccessMethod    string    `json:"access_method"`
	Status          string    `json:"status"`
	UpdateFrequency string    `json:"update_frequency"`
	LastCheckedAt   time.Time `json:"last_checked_at"`
}

type SourcesRepository struct {
	db *sql.DB
}

func NewSourcesRepository(store *Postgres) *SourcesRepository {
	return &SourcesRepository{db: store.DB()}
}

func (r *SourcesRepository) List(ctx context.Context) ([]SourceRecord, error) {
	const query = `
		SELECT id, name, base_url, source_type, access_method, status, update_frequency, last_checked_at
		FROM sources
		ORDER BY source_type, name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()

	items := make([]SourceRecord, 0)
	for rows.Next() {
		var item SourceRecord
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.BaseURL,
			&item.SourceType,
			&item.AccessMethod,
			&item.Status,
			&item.UpdateFrequency,
			&item.LastCheckedAt,
		); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sources: %w", err)
	}

	return items, nil
}

func (r *SourcesRepository) GetByBaseURL(ctx context.Context, baseURL string) (SourceRecord, error) {
	const query = `
		SELECT id, name, base_url, source_type, access_method, status, update_frequency, last_checked_at
		FROM sources
		WHERE base_url = $1
	`

	var item SourceRecord
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(baseURL)).Scan(
		&item.ID,
		&item.Name,
		&item.BaseURL,
		&item.SourceType,
		&item.AccessMethod,
		&item.Status,
		&item.UpdateFrequency,
		&item.LastCheckedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SourceRecord{}, ErrSourceNotFound
		}
		return SourceRecord{}, fmt.Errorf("get source by base url: %w", err)
	}

	return item, nil
}
