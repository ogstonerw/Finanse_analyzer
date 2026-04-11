package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrForecastNotFound = errors.New("forecast not found")

type ForecastRecord struct {
	ID                       string
	AssetID                  string
	AssetTicker              string
	AssetName                string
	EventID                  sql.NullString
	EventType                sql.NullString
	EventSummary             sql.NullString
	ForecastHorizon          string
	ForecastTime             time.Time
	DirectionLabel           string
	SignalStrength           float64
	ConfidenceScore          float64
	Explanation              string
	AIMode                   string
	ModelName                string
	MarketContextLabel       string
	MarketContextScore       float64
	MarketContextExplanation string
	KeyFactorsJSON           json.RawMessage
	PreparedRequestJSON      json.RawMessage
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type CreateForecastParams struct {
	AssetID                  string
	EventID                  *string
	ForecastHorizon          string
	ForecastTime             time.Time
	DirectionLabel           string
	SignalStrength           float64
	ConfidenceScore          float64
	Explanation              string
	AIMode                   string
	ModelName                string
	MarketContextLabel       string
	MarketContextScore       float64
	MarketContextExplanation string
	KeyFactorsJSON           json.RawMessage
	PreparedRequestJSON      json.RawMessage
}

type ForecastsRepository struct {
	db *sql.DB
}

func NewForecastsRepository(store *Postgres) *ForecastsRepository {
	return &ForecastsRepository{db: store.DB()}
}

func (r *ForecastsRepository) Create(ctx context.Context, params CreateForecastParams) (ForecastRecord, error) {
	const query = `
		WITH inserted AS (
			INSERT INTO forecasts (
				asset_id,
				event_id,
				forecast_horizon,
				forecast_time,
				direction_label,
				signal_strength,
				confidence_score,
				explanation,
				ai_mode,
				model_name,
				market_context_label,
				market_context_score,
				market_context_explanation,
				key_factors_json,
				prepared_request_json
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			RETURNING
				id,
				asset_id,
				event_id,
				forecast_horizon,
				forecast_time,
				direction_label,
				signal_strength,
				confidence_score,
				explanation,
				ai_mode,
				model_name,
				market_context_label,
				market_context_score,
				market_context_explanation,
				key_factors_json,
				prepared_request_json,
				created_at,
				updated_at
		)
		SELECT
			i.id,
			i.asset_id,
			a.ticker,
			a.name,
			i.event_id,
			e.event_type,
			e.summary,
			i.forecast_horizon,
			i.forecast_time,
			i.direction_label,
			i.signal_strength,
			i.confidence_score,
			i.explanation,
			i.ai_mode,
			i.model_name,
			i.market_context_label,
			i.market_context_score,
			i.market_context_explanation,
			i.key_factors_json,
			i.prepared_request_json,
			i.created_at,
			i.updated_at
		FROM inserted i
		INNER JOIN assets a ON a.id = i.asset_id
		LEFT JOIN events e ON e.id = i.event_id
	`

	var item ForecastRecord
	err := r.db.QueryRowContext(
		ctx,
		query,
		params.AssetID,
		params.EventID,
		params.ForecastHorizon,
		params.ForecastTime,
		params.DirectionLabel,
		params.SignalStrength,
		params.ConfidenceScore,
		params.Explanation,
		params.AIMode,
		params.ModelName,
		params.MarketContextLabel,
		params.MarketContextScore,
		params.MarketContextExplanation,
		jsonValueOrEmptyArray(params.KeyFactorsJSON),
		jsonValueOrNil(params.PreparedRequestJSON),
	).Scan(
		&item.ID,
		&item.AssetID,
		&item.AssetTicker,
		&item.AssetName,
		&item.EventID,
		&item.EventType,
		&item.EventSummary,
		&item.ForecastHorizon,
		&item.ForecastTime,
		&item.DirectionLabel,
		&item.SignalStrength,
		&item.ConfidenceScore,
		&item.Explanation,
		&item.AIMode,
		&item.ModelName,
		&item.MarketContextLabel,
		&item.MarketContextScore,
		&item.MarketContextExplanation,
		&item.KeyFactorsJSON,
		&item.PreparedRequestJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return ForecastRecord{}, fmt.Errorf("create forecast: %w", err)
	}

	return item, nil
}

func (r *ForecastsRepository) GetLatest(ctx context.Context) (ForecastRecord, error) {
	const query = `
		SELECT
			f.id,
			f.asset_id,
			a.ticker,
			a.name,
			f.event_id,
			e.event_type,
			e.summary,
			f.forecast_horizon,
			f.forecast_time,
			f.direction_label,
			f.signal_strength,
			f.confidence_score,
			f.explanation,
			f.ai_mode,
			f.model_name,
			f.market_context_label,
			f.market_context_score,
			f.market_context_explanation,
			f.key_factors_json,
			f.prepared_request_json,
			f.created_at,
			f.updated_at
		FROM forecasts f
		INNER JOIN assets a ON a.id = f.asset_id
		LEFT JOIN events e ON e.id = f.event_id
		ORDER BY f.forecast_time DESC, f.created_at DESC
		LIMIT 1
	`

	var item ForecastRecord
	err := r.db.QueryRowContext(ctx, query).Scan(
		&item.ID,
		&item.AssetID,
		&item.AssetTicker,
		&item.AssetName,
		&item.EventID,
		&item.EventType,
		&item.EventSummary,
		&item.ForecastHorizon,
		&item.ForecastTime,
		&item.DirectionLabel,
		&item.SignalStrength,
		&item.ConfidenceScore,
		&item.Explanation,
		&item.AIMode,
		&item.ModelName,
		&item.MarketContextLabel,
		&item.MarketContextScore,
		&item.MarketContextExplanation,
		&item.KeyFactorsJSON,
		&item.PreparedRequestJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ForecastRecord{}, ErrForecastNotFound
		}
		return ForecastRecord{}, fmt.Errorf("get latest forecast: %w", err)
	}

	return item, nil
}

func (r *ForecastsRepository) ListRecent(ctx context.Context, limit int) ([]ForecastRecord, error) {
	if limit <= 0 {
		limit = 5
	}

	const query = `
		SELECT
			f.id,
			f.asset_id,
			a.ticker,
			a.name,
			f.event_id,
			e.event_type,
			e.summary,
			f.forecast_horizon,
			f.forecast_time,
			f.direction_label,
			f.signal_strength,
			f.confidence_score,
			f.explanation,
			f.ai_mode,
			f.model_name,
			f.market_context_label,
			f.market_context_score,
			f.market_context_explanation,
			f.key_factors_json,
			f.prepared_request_json,
			f.created_at,
			f.updated_at
		FROM forecasts f
		INNER JOIN assets a ON a.id = f.asset_id
		LEFT JOIN events e ON e.id = f.event_id
		ORDER BY f.forecast_time DESC, f.created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent forecasts: %w", err)
	}
	defer rows.Close()

	items := make([]ForecastRecord, 0, limit)
	for rows.Next() {
		var item ForecastRecord
		if err := rows.Scan(
			&item.ID,
			&item.AssetID,
			&item.AssetTicker,
			&item.AssetName,
			&item.EventID,
			&item.EventType,
			&item.EventSummary,
			&item.ForecastHorizon,
			&item.ForecastTime,
			&item.DirectionLabel,
			&item.SignalStrength,
			&item.ConfidenceScore,
			&item.Explanation,
			&item.AIMode,
			&item.ModelName,
			&item.MarketContextLabel,
			&item.MarketContextScore,
			&item.MarketContextExplanation,
			&item.KeyFactorsJSON,
			&item.PreparedRequestJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan recent forecast: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent forecasts: %w", err)
	}

	return items, nil
}

func jsonValueOrEmptyArray(value json.RawMessage) any {
	if len(value) == 0 {
		return "[]"
	}

	return string(value)
}

func jsonValueOrNil(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}

	return string(value)
}
