package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"diploma-market-ai/02_product/backend/internal/regime"
	"diploma-market-ai/02_product/backend/internal/storage"
)

const dashboardIndicatorTimeframe = "1d"

type DashboardHandler struct {
	regimeService        *regime.Service
	assetsRepository     *storage.AssetsRepository
	eventsRepository     *storage.EventsRepository
	forecastsRepository  *storage.ForecastsRepository
	indicatorsRepository *storage.TechnicalIndicatorsRepository
}

type DashboardSummary struct {
	GeneratedAt     time.Time           `json:"generated_at"`
	Regime          regime.MarketRegime `json:"regime"`
	Assets          []DashboardAsset    `json:"assets"`
	LatestForecasts []DashboardForecast `json:"latest_forecasts"`
	RecentEvents    []DashboardEvent    `json:"recent_events"`
	Summary         string              `json:"summary"`
}

type DashboardAsset struct {
	ID              string              `json:"id"`
	Ticker          string              `json:"ticker"`
	Name            string              `json:"name"`
	AssetType       string              `json:"asset_type"`
	Sector          string              `json:"sector"`
	Currency        string              `json:"currency"`
	IsActive        bool                `json:"is_active"`
	LatestIndicator *DashboardIndicator `json:"latest_indicator,omitempty"`
}

type DashboardIndicator struct {
	IndicatorTime     time.Time `json:"indicator_time"`
	WeeklyReturn      *float64  `json:"weekly_return,omitempty"`
	RSI               *float64  `json:"rsi,omitempty"`
	Volatility        *float64  `json:"volatility,omitempty"`
	TrendDirection    *string   `json:"trend_direction,omitempty"`
	ChannelPosition   *float64  `json:"channel_position,omitempty"`
	CalculationStatus string    `json:"calculation_status"`
}

type DashboardForecast struct {
	ID                 string    `json:"id"`
	AssetTicker        string    `json:"asset_ticker"`
	AssetName          string    `json:"asset_name"`
	Horizon            string    `json:"horizon"`
	Direction          string    `json:"direction"`
	Strength           float64   `json:"strength"`
	Confidence         float64   `json:"confidence"`
	Explanation        string    `json:"explanation"`
	GeneratedAt        time.Time `json:"generated_at"`
	MarketContextLabel string    `json:"market_context_label"`
	MarketContextScore float64   `json:"market_context_score"`
	EventType          *string   `json:"event_type,omitempty"`
	EventSummary       *string   `json:"event_summary,omitempty"`
}

type DashboardEvent struct {
	ID          string    `json:"id"`
	PublishedAt time.Time `json:"published_at"`
	NewsTitle   string    `json:"news_title"`
	EventType   string    `json:"event_type"`
	Summary     string    `json:"summary"`
	AssetTicker *string   `json:"asset_ticker,omitempty"`
	AssetName   *string   `json:"asset_name,omitempty"`
}

func NewDashboardHandler(store *storage.Postgres, regimeService *regime.Service) *DashboardHandler {
	return &DashboardHandler{
		regimeService:        regimeService,
		assetsRepository:     storage.NewAssetsRepository(store),
		eventsRepository:     storage.NewEventsRepository(store),
		forecastsRepository:  storage.NewForecastsRepository(store),
		indicatorsRepository: storage.NewTechnicalIndicatorsRepository(store),
	}
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	item, err := h.buildSummary(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *DashboardHandler) buildSummary(ctx context.Context) (DashboardSummary, error) {
	currentRegime, err := h.regimeService.Current(ctx)
	if err != nil {
		return DashboardSummary{}, err
	}

	assets, weakAssets, err := h.loadAssets(ctx)
	if err != nil {
		return DashboardSummary{}, err
	}

	forecasts, err := h.loadForecasts(ctx)
	if err != nil {
		return DashboardSummary{}, err
	}

	events, err := h.loadEvents(ctx)
	if err != nil {
		return DashboardSummary{}, err
	}

	return DashboardSummary{
		GeneratedAt:     currentRegime.CalculatedAt,
		Regime:          currentRegime,
		Assets:          assets,
		LatestForecasts: forecasts,
		RecentEvents:    events,
		Summary:         buildDashboardSummaryText(currentRegime, weakAssets, len(assets), forecasts, events),
	}, nil
}

func (h *DashboardHandler) loadAssets(ctx context.Context) ([]DashboardAsset, int, error) {
	records, err := h.assetsRepository.List(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]DashboardAsset, 0, len(records))
	weakAssets := 0

	for _, record := range records {
		if !record.IsActive {
			continue
		}

		item := DashboardAsset{
			ID:        record.ID,
			Ticker:    record.Ticker,
			Name:      record.Name,
			AssetType: record.AssetType,
			Sector:    record.Sector,
			Currency:  record.Currency,
			IsActive:  record.IsActive,
		}

		indicator, err := h.indicatorsRepository.GetLatestByAsset(ctx, record.ID, dashboardIndicatorTimeframe)
		switch {
		case err == nil:
			item.LatestIndicator = &DashboardIndicator{
				IndicatorTime:     indicator.IndicatorTime,
				WeeklyReturn:      nullFloat64ToPointer(indicator.WeeklyReturn),
				RSI:               nullFloat64ToPointer(indicator.RSI),
				Volatility:        nullFloat64ToPointer(indicator.Volatility),
				TrendDirection:    nullStringToPointer(indicator.TrendDirection),
				ChannelPosition:   nullFloat64ToPointer(indicator.ChannelPosition),
				CalculationStatus: indicator.CalculationStatus,
			}
			if indicator.WeeklyReturn.Valid && indicator.WeeklyReturn.Float64 < 0 {
				weakAssets++
			}
		case errors.Is(err, storage.ErrTechnicalIndicatorNotFound):
		default:
			return nil, 0, fmt.Errorf("load latest indicator for %s: %w", record.Ticker, err)
		}

		items = append(items, item)
	}

	return items, weakAssets, nil
}

func (h *DashboardHandler) loadForecasts(ctx context.Context) ([]DashboardForecast, error) {
	records, err := h.forecastsRepository.ListRecent(ctx, 5)
	if err != nil {
		return nil, err
	}

	items := make([]DashboardForecast, 0, len(records))
	for _, record := range records {
		items = append(items, DashboardForecast{
			ID:                 record.ID,
			AssetTicker:        record.AssetTicker,
			AssetName:          record.AssetName,
			Horizon:            record.ForecastHorizon,
			Direction:          record.DirectionLabel,
			Strength:           record.SignalStrength,
			Confidence:         record.ConfidenceScore,
			Explanation:        record.Explanation,
			GeneratedAt:        record.ForecastTime,
			MarketContextLabel: record.MarketContextLabel,
			MarketContextScore: record.MarketContextScore,
			EventType:          nullStringToPointer(record.EventType),
			EventSummary:       nullStringToPointer(record.EventSummary),
		})
	}

	return items, nil
}

func (h *DashboardHandler) loadEvents(ctx context.Context) ([]DashboardEvent, error) {
	records, err := h.eventsRepository.ListRecent(ctx, 5)
	if err != nil {
		return nil, err
	}

	items := make([]DashboardEvent, 0, len(records))
	for _, record := range records {
		items = append(items, DashboardEvent{
			ID:          record.ID,
			PublishedAt: eventTime(record),
			NewsTitle:   record.NewsTitle,
			EventType:   record.EventType,
			Summary:     record.Summary,
			AssetTicker: nullStringToPointer(record.AssetTicker),
			AssetName:   nullStringToPointer(record.AssetName),
		})
	}

	return items, nil
}

func buildDashboardSummaryText(
	currentRegime regime.MarketRegime,
	weakAssets int,
	totalAssets int,
	forecasts []DashboardForecast,
	events []DashboardEvent,
) string {
	parts := []string{
		fmt.Sprintf(
			"Current market regime is %s with score %.2f under the temporary rule-based MVP crisisometer.",
			currentRegime.RegimeLabel,
			currentRegime.RegimeScore,
		),
	}

	if totalAssets > 0 {
		parts = append(parts, fmt.Sprintf("%d of %d tracked assets show negative weekly dynamics.", weakAssets, totalAssets))
	}

	if len(forecasts) > 0 {
		parts = append(parts, fmt.Sprintf("Latest forecast set contains %d recent records, most recent for %s.", len(forecasts), forecasts[0].AssetTicker))
	}

	if len(events) > 0 {
		eventLabel := events[0].EventType
		if events[0].AssetTicker != nil && strings.TrimSpace(*events[0].AssetTicker) != "" {
			eventLabel = fmt.Sprintf("%s for %s", eventLabel, *events[0].AssetTicker)
		}
		parts = append(parts, fmt.Sprintf("Latest event focus: %s.", eventLabel))
	}

	return strings.Join(parts, " ")
}

func eventTime(item storage.EventRecord) time.Time {
	if !item.PublishedAt.IsZero() {
		return item.PublishedAt
	}
	return item.ExtractedAt
}

func nullStringToPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}

func nullFloat64ToPointer(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	result := value.Float64
	return &result
}
