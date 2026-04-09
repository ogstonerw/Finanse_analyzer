package prices

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"diploma-market-ai/02_product/backend/internal/collectors"
	"diploma-market-ai/02_product/backend/internal/storage"
)

const dailyTimeframe = "1d"

var ErrAssetNotFound = errors.New("asset not found")

type Service struct {
	assetsRepository *storage.AssetsRepository
	pricesRepository *storage.PriceCandlesRepository
	collector        *collectors.MOEXCollector
}

type Candle struct {
	Timeframe  string    `json:"timeframe"`
	CandleTime time.Time `json:"candle_time"`
	OpenPrice  float64   `json:"open_price"`
	HighPrice  float64   `json:"high_price"`
	LowPrice   float64   `json:"low_price"`
	ClosePrice float64   `json:"close_price"`
	Volume     float64   `json:"volume"`
}

func NewService(store *storage.Postgres, collector *collectors.MOEXCollector) *Service {
	return &Service{
		assetsRepository: storage.NewAssetsRepository(store),
		pricesRepository: storage.NewPriceCandlesRepository(store),
		collector:        collector,
	}
}

func (s *Service) SyncSupportedDailyCandles(ctx context.Context) error {
	if s.collector == nil {
		return errors.New("prices collector is not configured")
	}

	var failed []string
	for _, ticker := range s.collector.SupportedTickers() {
		if err := s.syncTicker(ctx, ticker); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", ticker, err))
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("sync daily candles failed for %s", strings.Join(failed, "; "))
	}

	return nil
}

func (s *Service) ListByTicker(ctx context.Context, ticker string) ([]Candle, error) {
	asset, err := s.assetsRepository.GetByTicker(ctx, ticker)
	if err != nil {
		if errors.Is(err, storage.ErrAssetNotFound) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	items, err := s.pricesRepository.ListByAsset(ctx, asset.ID, dailyTimeframe)
	if err != nil {
		return nil, err
	}

	result := make([]Candle, 0, len(items))
	for _, item := range items {
		result = append(result, Candle{
			Timeframe:  item.Timeframe,
			CandleTime: item.CandleTime,
			OpenPrice:  item.OpenPrice,
			HighPrice:  item.HighPrice,
			LowPrice:   item.LowPrice,
			ClosePrice: item.ClosePrice,
			Volume:     item.Volume,
		})
	}

	return result, nil
}

func (s *Service) syncTicker(ctx context.Context, ticker string) error {
	asset, err := s.assetsRepository.GetByTicker(ctx, ticker)
	if err != nil {
		return err
	}

	from, err := s.nextSyncFrom(ctx, asset.ID)
	if err != nil {
		return err
	}

	till := time.Now().UTC()
	if from.After(till) {
		return nil
	}

	candles, err := s.collector.FetchDailyCandles(ctx, ticker, from, till)
	if err != nil {
		return err
	}

	params := make([]storage.UpsertPriceCandleParams, 0, len(candles))
	for _, candle := range candles {
		params = append(params, storage.UpsertPriceCandleParams{
			AssetID:    asset.ID,
			Timeframe:  dailyTimeframe,
			CandleTime: candle.Time,
			OpenPrice:  candle.Open,
			HighPrice:  candle.High,
			LowPrice:   candle.Low,
			ClosePrice: candle.Close,
			Volume:     candle.Volume,
		})
	}

	return s.pricesRepository.UpsertBatch(ctx, params)
}

func (s *Service) nextSyncFrom(ctx context.Context, assetID string) (time.Time, error) {
	item, err := s.pricesRepository.GetLatestByAsset(ctx, assetID, dailyTimeframe)
	if err == nil {
		return item.CandleTime.AddDate(0, 0, 1).UTC(), nil
	}

	if errors.Is(err, storage.ErrPriceCandleNotFound) {
		return time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, err
}
