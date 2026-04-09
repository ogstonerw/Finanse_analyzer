package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrPriceCandleNotFound = errors.New("price candle not found")

type PriceCandleRecord struct {
	ID         string
	AssetID    string
	Timeframe  string
	CandleTime time.Time
	OpenPrice  float64
	HighPrice  float64
	LowPrice   float64
	ClosePrice float64
	Volume     float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UpsertPriceCandleParams struct {
	AssetID    string
	Timeframe  string
	CandleTime time.Time
	OpenPrice  float64
	HighPrice  float64
	LowPrice   float64
	ClosePrice float64
	Volume     float64
}

type PriceCandlesRepository struct {
	db *sql.DB
}

func NewPriceCandlesRepository(store *Postgres) *PriceCandlesRepository {
	return &PriceCandlesRepository{db: store.DB()}
}

func (r *PriceCandlesRepository) UpsertBatch(ctx context.Context, candles []UpsertPriceCandleParams) error {
	if len(candles) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin upsert price candles tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const query = `
		INSERT INTO price_candles (
			asset_id, timeframe, candle_time, open_price, high_price, low_price, close_price, volume
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (asset_id, timeframe, candle_time)
		DO UPDATE SET
			open_price = EXCLUDED.open_price,
			high_price = EXCLUDED.high_price,
			low_price = EXCLUDED.low_price,
			close_price = EXCLUDED.close_price,
			volume = EXCLUDED.volume,
			updated_at = NOW()
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare upsert price candles: %w", err)
	}
	defer stmt.Close()

	for _, candle := range candles {
		if _, err = stmt.ExecContext(
			ctx,
			candle.AssetID,
			candle.Timeframe,
			candle.CandleTime,
			candle.OpenPrice,
			candle.HighPrice,
			candle.LowPrice,
			candle.ClosePrice,
			candle.Volume,
		); err != nil {
			return fmt.Errorf("exec upsert price candle: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit upsert price candles: %w", err)
	}

	return nil
}

func (r *PriceCandlesRepository) ListByAsset(ctx context.Context, assetID, timeframe string) ([]PriceCandleRecord, error) {
	const query = `
		SELECT id, asset_id, timeframe, candle_time, open_price, high_price, low_price, close_price, volume, created_at, updated_at
		FROM price_candles
		WHERE asset_id = $1 AND timeframe = $2
		ORDER BY candle_time
	`

	rows, err := r.db.QueryContext(ctx, query, assetID, timeframe)
	if err != nil {
		return nil, fmt.Errorf("list price candles by asset: %w", err)
	}
	defer rows.Close()

	items := make([]PriceCandleRecord, 0)
	for rows.Next() {
		var item PriceCandleRecord
		if err := rows.Scan(
			&item.ID,
			&item.AssetID,
			&item.Timeframe,
			&item.CandleTime,
			&item.OpenPrice,
			&item.HighPrice,
			&item.LowPrice,
			&item.ClosePrice,
			&item.Volume,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan price candle: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate price candles: %w", err)
	}

	return items, nil
}

func (r *PriceCandlesRepository) GetLatestByAsset(ctx context.Context, assetID, timeframe string) (PriceCandleRecord, error) {
	const query = `
		SELECT id, asset_id, timeframe, candle_time, open_price, high_price, low_price, close_price, volume, created_at, updated_at
		FROM price_candles
		WHERE asset_id = $1 AND timeframe = $2
		ORDER BY candle_time DESC
		LIMIT 1
	`

	var item PriceCandleRecord
	err := r.db.QueryRowContext(ctx, query, assetID, timeframe).Scan(
		&item.ID,
		&item.AssetID,
		&item.Timeframe,
		&item.CandleTime,
		&item.OpenPrice,
		&item.HighPrice,
		&item.LowPrice,
		&item.ClosePrice,
		&item.Volume,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PriceCandleRecord{}, ErrPriceCandleNotFound
		}
		return PriceCandleRecord{}, fmt.Errorf("get latest price candle by asset: %w", err)
	}

	return item, nil
}
