package indicators

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"diploma-market-ai/02_product/backend/internal/storage"
)

const (
	dailyTimeframe     = "1d"
	rsiPeriod          = 14
	weeklyReturnWindow = 5
	volatilityWindow   = 20
	channelWindow      = 20
	trendShortWindow   = 5
	trendLongWindow    = 20
	statusInsufficient = "insufficient_data"
	statusPartial      = "partial"
	statusReady        = "ready"
	tradingDaysPerYear = 252.0
)

var ErrAssetNotFound = errors.New("asset not found")

type Service struct {
	assetsRepository     *storage.AssetsRepository
	pricesRepository     *storage.PriceCandlesRepository
	indicatorsRepository *storage.TechnicalIndicatorsRepository
}

type Indicator struct {
	IndicatorTime     time.Time `json:"indicator_time"`
	Timeframe         string    `json:"timeframe"`
	WeeklyReturn      *float64  `json:"weekly_return"`
	RSI               *float64  `json:"rsi"`
	Volatility        *float64  `json:"volatility"`
	TrendDirection    *string   `json:"trend_direction"`
	ChannelPosition   *float64  `json:"channel_position"`
	CalculationStatus string    `json:"calculation_status"`
}

func NewService(store *storage.Postgres) *Service {
	return &Service{
		assetsRepository:     storage.NewAssetsRepository(store),
		pricesRepository:     storage.NewPriceCandlesRepository(store),
		indicatorsRepository: storage.NewTechnicalIndicatorsRepository(store),
	}
}

func (s *Service) SyncAllDailyIndicators(ctx context.Context) error {
	assets, err := s.assetsRepository.List(ctx)
	if err != nil {
		return err
	}

	var failed []string
	for _, asset := range assets {
		if err := s.syncAsset(ctx, asset.ID); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", asset.Ticker, err))
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("sync technical indicators failed for %s", strings.Join(failed, "; "))
	}

	return nil
}

func (s *Service) ListByTicker(ctx context.Context, ticker string) ([]Indicator, error) {
	asset, err := s.assetsRepository.GetByTicker(ctx, ticker)
	if err != nil {
		if errors.Is(err, storage.ErrAssetNotFound) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	items, err := s.indicatorsRepository.ListByAsset(ctx, asset.ID, dailyTimeframe)
	if err != nil {
		return nil, err
	}

	result := make([]Indicator, 0, len(items))
	for _, item := range items {
		result = append(result, Indicator{
			IndicatorTime:     item.IndicatorTime,
			Timeframe:         item.Timeframe,
			WeeklyReturn:      nullFloat64ToPointer(item.WeeklyReturn),
			RSI:               nullFloat64ToPointer(item.RSI),
			Volatility:        nullFloat64ToPointer(item.Volatility),
			TrendDirection:    nullStringToPointer(item.TrendDirection),
			ChannelPosition:   nullFloat64ToPointer(item.ChannelPosition),
			CalculationStatus: item.CalculationStatus,
		})
	}

	return result, nil
}

func (s *Service) syncAsset(ctx context.Context, assetID string) error {
	candles, err := s.pricesRepository.ListByAsset(ctx, assetID, dailyTimeframe)
	if err != nil {
		return err
	}

	if len(candles) == 0 {
		return nil
	}

	items := make([]storage.UpsertTechnicalIndicatorParams, 0, len(candles))
	for index := range candles {
		calculated := calculateIndicator(candles, index)
		items = append(items, storage.UpsertTechnicalIndicatorParams{
			AssetID:           assetID,
			IndicatorTime:     candles[index].CandleTime,
			Timeframe:         dailyTimeframe,
			WeeklyReturn:      calculated.WeeklyReturn,
			RSI:               calculated.RSI,
			Volatility:        calculated.Volatility,
			TrendDirection:    calculated.TrendDirection,
			ChannelPosition:   calculated.ChannelPosition,
			CalculationStatus: calculated.CalculationStatus,
		})
	}

	return s.indicatorsRepository.UpsertBatch(ctx, items)
}

type calculatedIndicator struct {
	WeeklyReturn      *float64
	RSI               *float64
	Volatility        *float64
	TrendDirection    *string
	ChannelPosition   *float64
	CalculationStatus string
}

func calculateIndicator(candles []storage.PriceCandleRecord, index int) calculatedIndicator {
	weeklyReturn := calculateWeeklyReturn(candles, index)
	rsi := calculateRSI(candles, index)
	volatility := calculateVolatility(candles, index)
	trendDirection := calculateTrendDirection(candles, index)
	channelPosition := calculateChannelPosition(candles, index)

	status := deriveCalculationStatus(
		weeklyReturn != nil,
		rsi != nil,
		volatility != nil,
		trendDirection != nil,
		channelPosition != nil,
	)

	return calculatedIndicator{
		WeeklyReturn:      weeklyReturn,
		RSI:               rsi,
		Volatility:        volatility,
		TrendDirection:    trendDirection,
		ChannelPosition:   channelPosition,
		CalculationStatus: status,
	}
}

func calculateWeeklyReturn(candles []storage.PriceCandleRecord, index int) *float64 {
	if index < weeklyReturnWindow {
		return nil
	}

	baseClose := candles[index-weeklyReturnWindow].ClosePrice
	if baseClose == 0 {
		return nil
	}

	value := candles[index].ClosePrice/baseClose - 1
	return float64Pointer(value)
}

func calculateRSI(candles []storage.PriceCandleRecord, index int) *float64 {
	if index < rsiPeriod {
		return nil
	}

	start := index - rsiPeriod + 1
	var gains float64
	var losses float64

	for i := start; i <= index; i++ {
		delta := candles[i].ClosePrice - candles[i-1].ClosePrice
		switch {
		case delta > 0:
			gains += delta
		case delta < 0:
			losses += -delta
		}
	}

	avgGain := gains / float64(rsiPeriod)
	avgLoss := losses / float64(rsiPeriod)

	switch {
	case avgLoss == 0 && avgGain == 0:
		return float64Pointer(50)
	case avgLoss == 0:
		return float64Pointer(100)
	case avgGain == 0:
		return float64Pointer(0)
	}

	rs := avgGain / avgLoss
	value := 100 - (100 / (1 + rs))
	return float64Pointer(value)
}

func calculateVolatility(candles []storage.PriceCandleRecord, index int) *float64 {
	if index < volatilityWindow {
		return nil
	}

	start := index - volatilityWindow + 1
	returns := make([]float64, 0, volatilityWindow)
	for i := start; i <= index; i++ {
		prevClose := candles[i-1].ClosePrice
		currClose := candles[i].ClosePrice
		if prevClose <= 0 || currClose <= 0 {
			return nil
		}

		returns = append(returns, math.Log(currClose/prevClose))
	}

	if len(returns) < 2 {
		return nil
	}

	mean := average(returns)

	var variance float64
	for _, value := range returns {
		diff := value - mean
		variance += diff * diff
	}

	variance /= float64(len(returns) - 1)
	volatility := math.Sqrt(variance) * math.Sqrt(tradingDaysPerYear)

	return float64Pointer(volatility)
}

func calculateTrendDirection(candles []storage.PriceCandleRecord, index int) *string {
	if index < trendLongWindow-1 {
		return nil
	}

	shortMA := averageClose(candles[index-trendShortWindow+1 : index+1])
	longMA := averageClose(candles[index-trendLongWindow+1 : index+1])

	switch {
	case shortMA > longMA:
		return stringPointer("up")
	case shortMA < longMA:
		return stringPointer("down")
	default:
		return stringPointer("flat")
	}
}

func calculateChannelPosition(candles []storage.PriceCandleRecord, index int) *float64 {
	if index < channelWindow-1 {
		return nil
	}

	window := candles[index-channelWindow+1 : index+1]
	minLow := window[0].LowPrice
	maxHigh := window[0].HighPrice

	for _, candle := range window[1:] {
		if candle.LowPrice < minLow {
			minLow = candle.LowPrice
		}
		if candle.HighPrice > maxHigh {
			maxHigh = candle.HighPrice
		}
	}

	if maxHigh <= minLow {
		return nil
	}

	position := (candles[index].ClosePrice - minLow) / (maxHigh - minLow)
	if position < 0 {
		position = 0
	}
	if position > 1 {
		position = 1
	}

	return float64Pointer(position)
}

func deriveCalculationStatus(flags ...bool) string {
	var available int
	for _, flag := range flags {
		if flag {
			available++
		}
	}

	switch {
	case available == 0:
		return statusInsufficient
	case available == len(flags):
		return statusReady
	default:
		return statusPartial
	}
}

func average(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func averageClose(candles []storage.PriceCandleRecord) float64 {
	var total float64
	for _, candle := range candles {
		total += candle.ClosePrice
	}
	return total / float64(len(candles))
}

func float64Pointer(value float64) *float64 {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

func nullFloat64ToPointer(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	result := value.Float64
	return &result
}

func nullStringToPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}
