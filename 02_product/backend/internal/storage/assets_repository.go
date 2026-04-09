package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrAssetNotFound = errors.New("asset not found")

type AssetRecord struct {
	ID        string
	Ticker    string
	Name      string
	AssetType string
	Sector    string
	Currency  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AssetsRepository struct {
	db *sql.DB
}

func NewAssetsRepository(store *Postgres) *AssetsRepository {
	return &AssetsRepository{db: store.DB()}
}

func (r *AssetsRepository) List(ctx context.Context) ([]AssetRecord, error) {
	const query = `
		SELECT id, ticker, name, asset_type, sector, currency, is_active, created_at, updated_at
		FROM assets
		ORDER BY ticker
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	items := make([]AssetRecord, 0)
	for rows.Next() {
		var item AssetRecord
		if err := rows.Scan(
			&item.ID,
			&item.Ticker,
			&item.Name,
			&item.AssetType,
			&item.Sector,
			&item.Currency,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assets: %w", err)
	}

	return items, nil
}

func (r *AssetsRepository) GetByTicker(ctx context.Context, ticker string) (AssetRecord, error) {
	const query = `
		SELECT id, ticker, name, asset_type, sector, currency, is_active, created_at, updated_at
		FROM assets
		WHERE ticker = $1
	`

	normalizedTicker := strings.ToUpper(strings.TrimSpace(ticker))

	var item AssetRecord
	err := r.db.QueryRowContext(ctx, query, normalizedTicker).Scan(
		&item.ID,
		&item.Ticker,
		&item.Name,
		&item.AssetType,
		&item.Sector,
		&item.Currency,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AssetRecord{}, ErrAssetNotFound
		}
		return AssetRecord{}, fmt.Errorf("get asset by ticker: %w", err)
	}

	return item, nil
}
