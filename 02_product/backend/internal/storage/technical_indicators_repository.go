package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TechnicalIndicatorRecord struct {
	ID                string
	AssetID           string
	IndicatorTime     time.Time
	Timeframe         string
	WeeklyReturn      sql.NullFloat64
	RSI               sql.NullFloat64
	Volatility        sql.NullFloat64
	TrendDirection    sql.NullString
	ChannelPosition   sql.NullFloat64
	CalculationStatus string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UpsertTechnicalIndicatorParams struct {
	AssetID           string
	IndicatorTime     time.Time
	Timeframe         string
	WeeklyReturn      *float64
	RSI               *float64
	Volatility        *float64
	TrendDirection    *string
	ChannelPosition   *float64
	CalculationStatus string
}

type TechnicalIndicatorsRepository struct {
	db *sql.DB
}

func NewTechnicalIndicatorsRepository(store *Postgres) *TechnicalIndicatorsRepository {
	return &TechnicalIndicatorsRepository{db: store.DB()}
}

func (r *TechnicalIndicatorsRepository) UpsertBatch(ctx context.Context, items []UpsertTechnicalIndicatorParams) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin upsert technical indicators tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const query = `
		INSERT INTO technical_indicators (
			asset_id, indicator_time, timeframe, weekly_return, rsi, volatility, trend_direction, channel_position, calculation_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (asset_id, timeframe, indicator_time)
		DO UPDATE SET
			weekly_return = EXCLUDED.weekly_return,
			rsi = EXCLUDED.rsi,
			volatility = EXCLUDED.volatility,
			trend_direction = EXCLUDED.trend_direction,
			channel_position = EXCLUDED.channel_position,
			calculation_status = EXCLUDED.calculation_status,
			updated_at = NOW()
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare upsert technical indicators: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err = stmt.ExecContext(
			ctx,
			item.AssetID,
			item.IndicatorTime,
			item.Timeframe,
			item.WeeklyReturn,
			item.RSI,
			item.Volatility,
			item.TrendDirection,
			item.ChannelPosition,
			item.CalculationStatus,
		); err != nil {
			return fmt.Errorf("exec upsert technical indicator: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit upsert technical indicators: %w", err)
	}

	return nil
}

func (r *TechnicalIndicatorsRepository) ListByAsset(ctx context.Context, assetID, timeframe string) ([]TechnicalIndicatorRecord, error) {
	const query = `
		SELECT
			id,
			asset_id,
			indicator_time,
			timeframe,
			weekly_return,
			rsi,
			volatility,
			trend_direction,
			channel_position,
			calculation_status,
			created_at,
			updated_at
		FROM technical_indicators
		WHERE asset_id = $1 AND timeframe = $2
		ORDER BY indicator_time
	`

	rows, err := r.db.QueryContext(ctx, query, assetID, timeframe)
	if err != nil {
		return nil, fmt.Errorf("list technical indicators by asset: %w", err)
	}
	defer rows.Close()

	items := make([]TechnicalIndicatorRecord, 0)
	for rows.Next() {
		var item TechnicalIndicatorRecord
		if err := rows.Scan(
			&item.ID,
			&item.AssetID,
			&item.IndicatorTime,
			&item.Timeframe,
			&item.WeeklyReturn,
			&item.RSI,
			&item.Volatility,
			&item.TrendDirection,
			&item.ChannelPosition,
			&item.CalculationStatus,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan technical indicator: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate technical indicators: %w", err)
	}

	return items, nil
}
