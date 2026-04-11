package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrMarketRegimeNotFound = errors.New("market regime not found")

type MarketRegimeRecord struct {
	ID                   string
	AssetID              sql.NullString
	RegimeTime           time.Time
	RegimeLabel          string
	RegimeScore          float64
	MarketStressScore    float64
	NewsStressScore      float64
	MacroStressScore     float64
	CommodityStressScore float64
	BreadthStressScore   float64
	Summary              string
	Explanation          string
	CalculationModel     string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type SaveMarketRegimeParams struct {
	AssetID              *string
	RegimeTime           time.Time
	RegimeLabel          string
	RegimeScore          float64
	MarketStressScore    float64
	NewsStressScore      float64
	MacroStressScore     float64
	CommodityStressScore float64
	BreadthStressScore   float64
	Summary              string
	Explanation          string
	CalculationModel     string
}

type MarketRegimesRepository struct {
	db *sql.DB
}

type rowScanner interface {
	Scan(dest ...any) error
}

func NewMarketRegimesRepository(store *Postgres) *MarketRegimesRepository {
	return &MarketRegimesRepository{db: store.DB()}
}

func (r *MarketRegimesRepository) Save(ctx context.Context, params SaveMarketRegimeParams) (MarketRegimeRecord, error) {
	if params.CalculationModel == "" {
		params.CalculationModel = "rule_based_mvp"
	}

	const lookupQuery = `
		SELECT id
		FROM market_regime
		WHERE asset_id IS NOT DISTINCT FROM $1
			AND regime_time = $2
			AND calculation_model = $3
		LIMIT 1
	`

	var existingID string
	err := r.db.QueryRowContext(
		ctx,
		lookupQuery,
		nullableStringValue(params.AssetID),
		params.RegimeTime,
		params.CalculationModel,
	).Scan(&existingID)
	switch {
	case err == nil:
		return r.update(ctx, existingID, params)
	case errors.Is(err, sql.ErrNoRows):
		return r.create(ctx, params)
	default:
		return MarketRegimeRecord{}, fmt.Errorf("lookup market regime snapshot: %w", err)
	}
}

func (r *MarketRegimesRepository) GetLatest(ctx context.Context, assetID *string) (MarketRegimeRecord, error) {
	query := `
		SELECT
			id,
			asset_id,
			regime_time,
			regime_label,
			regime_score,
			market_stress_score,
			news_stress_score,
			macro_stress_score,
			commodity_stress_score,
			breadth_stress_score,
			summary,
			explanation,
			calculation_model,
			created_at,
			updated_at
		FROM market_regime
		WHERE asset_id IS NULL
		ORDER BY regime_time DESC, updated_at DESC
		LIMIT 1
	`

	var row rowScanner
	if assetID != nil {
		query = `
			SELECT
				id,
				asset_id,
				regime_time,
				regime_label,
				regime_score,
				market_stress_score,
				news_stress_score,
				macro_stress_score,
				commodity_stress_score,
				breadth_stress_score,
				summary,
				explanation,
				calculation_model,
				created_at,
				updated_at
			FROM market_regime
			WHERE asset_id = $1
			ORDER BY regime_time DESC, updated_at DESC
			LIMIT 1
		`
		row = r.db.QueryRowContext(ctx, query, *assetID)
	} else {
		row = r.db.QueryRowContext(ctx, query)
	}

	item, err := scanMarketRegime(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return MarketRegimeRecord{}, ErrMarketRegimeNotFound
		}
		return MarketRegimeRecord{}, fmt.Errorf("get latest market regime: %w", err)
	}

	return item, nil
}

func (r *MarketRegimesRepository) create(ctx context.Context, params SaveMarketRegimeParams) (MarketRegimeRecord, error) {
	const query = `
		INSERT INTO market_regime (
			asset_id,
			regime_time,
			regime_label,
			regime_score,
			market_stress_score,
			news_stress_score,
			macro_stress_score,
			commodity_stress_score,
			breadth_stress_score,
			summary,
			explanation,
			calculation_model
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING
			id,
			asset_id,
			regime_time,
			regime_label,
			regime_score,
			market_stress_score,
			news_stress_score,
			macro_stress_score,
			commodity_stress_score,
			breadth_stress_score,
			summary,
			explanation,
			calculation_model,
			created_at,
			updated_at
	`

	item, err := scanMarketRegime(r.db.QueryRowContext(
		ctx,
		query,
		nullableStringValue(params.AssetID),
		params.RegimeTime,
		params.RegimeLabel,
		params.RegimeScore,
		params.MarketStressScore,
		params.NewsStressScore,
		params.MacroStressScore,
		params.CommodityStressScore,
		params.BreadthStressScore,
		params.Summary,
		params.Explanation,
		params.CalculationModel,
	))
	if err != nil {
		return MarketRegimeRecord{}, fmt.Errorf("create market regime snapshot: %w", err)
	}

	return item, nil
}

func (r *MarketRegimesRepository) update(ctx context.Context, id string, params SaveMarketRegimeParams) (MarketRegimeRecord, error) {
	const query = `
		UPDATE market_regime
		SET
			asset_id = $2,
			regime_label = $3,
			regime_score = $4,
			market_stress_score = $5,
			news_stress_score = $6,
			macro_stress_score = $7,
			commodity_stress_score = $8,
			breadth_stress_score = $9,
			summary = $10,
			explanation = $11,
			calculation_model = $12,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			asset_id,
			regime_time,
			regime_label,
			regime_score,
			market_stress_score,
			news_stress_score,
			macro_stress_score,
			commodity_stress_score,
			breadth_stress_score,
			summary,
			explanation,
			calculation_model,
			created_at,
			updated_at
	`

	item, err := scanMarketRegime(r.db.QueryRowContext(
		ctx,
		query,
		id,
		nullableStringValue(params.AssetID),
		params.RegimeLabel,
		params.RegimeScore,
		params.MarketStressScore,
		params.NewsStressScore,
		params.MacroStressScore,
		params.CommodityStressScore,
		params.BreadthStressScore,
		params.Summary,
		params.Explanation,
		params.CalculationModel,
	))
	if err != nil {
		return MarketRegimeRecord{}, fmt.Errorf("update market regime snapshot: %w", err)
	}

	return item, nil
}

func scanMarketRegime(scanner rowScanner) (MarketRegimeRecord, error) {
	var item MarketRegimeRecord
	err := scanner.Scan(
		&item.ID,
		&item.AssetID,
		&item.RegimeTime,
		&item.RegimeLabel,
		&item.RegimeScore,
		&item.MarketStressScore,
		&item.NewsStressScore,
		&item.MacroStressScore,
		&item.CommodityStressScore,
		&item.BreadthStressScore,
		&item.Summary,
		&item.Explanation,
		&item.CalculationModel,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return MarketRegimeRecord{}, err
	}

	return item, nil
}

func nullableStringValue(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}
