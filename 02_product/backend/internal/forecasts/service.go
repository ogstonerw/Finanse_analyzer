package forecasts

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"diploma-market-ai/02_product/backend/internal/ai"
	"diploma-market-ai/02_product/backend/internal/storage"
)

const (
	dailyTimeframe           = "1d"
	defaultForecastHorizon   = "1w"
	marketContextAssetTicker = "IMOEX"
)

var (
	ErrAssetNotFound      = errors.New("asset not found")
	ErrEventNotFound      = errors.New("event not found")
	ErrForecastNotFound   = errors.New("forecast not found")
	ErrEventAssetMismatch = errors.New("event does not match asset")
)

type Service struct {
	assetsRepository     *storage.AssetsRepository
	eventsRepository     *storage.EventsRepository
	newsRepository       *storage.NewsItemsRepository
	indicatorsRepository *storage.TechnicalIndicatorsRepository
	forecastsRepository  *storage.ForecastsRepository
	aiClient             *ai.Client
}

type GenerateRequest struct {
	Ticker  string `json:"ticker"`
	EventID string `json:"event_id,omitempty"`
	Horizon string `json:"horizon,omitempty"`
}

type Forecast struct {
	ID              string          `json:"id"`
	AssetID         string          `json:"asset_id"`
	AssetTicker     string          `json:"asset_ticker"`
	AssetName       string          `json:"asset_name"`
	EventID         *string         `json:"event_id,omitempty"`
	EventType       *string         `json:"event_type,omitempty"`
	EventSummary    *string         `json:"event_summary,omitempty"`
	Horizon         string          `json:"horizon"`
	Direction       string          `json:"direction"`
	Strength        float64         `json:"strength"`
	Confidence      float64         `json:"confidence"`
	Explanation     string          `json:"explanation"`
	GeneratedAt     time.Time       `json:"generated_at"`
	AIMode          string          `json:"ai_mode"`
	Model           string          `json:"model"`
	MarketContext   MarketContext   `json:"market_context"`
	KeyFactors      []string        `json:"key_factors"`
	PreparedRequest json.RawMessage `json:"prepared_request,omitempty"`
}

type MarketContext struct {
	Label       string  `json:"label"`
	Score       float64 `json:"score"`
	Explanation string  `json:"explanation"`
}

type derivedMarketContext struct {
	Label       string
	Score       float64
	Explanation string
}

func NewService(store *storage.Postgres, aiClient *ai.Client) *Service {
	if aiClient == nil {
		aiClient = ai.NewClient(ai.Config{Mode: ai.ModeFallback})
	}

	return &Service{
		assetsRepository:     storage.NewAssetsRepository(store),
		eventsRepository:     storage.NewEventsRepository(store),
		newsRepository:       storage.NewNewsItemsRepository(store),
		indicatorsRepository: storage.NewTechnicalIndicatorsRepository(store),
		forecastsRepository:  storage.NewForecastsRepository(store),
		aiClient:             aiClient,
	}
}

func (s *Service) Generate(ctx context.Context, req GenerateRequest) (Forecast, error) {
	asset, err := s.assetsRepository.GetByTicker(ctx, req.Ticker)
	if err != nil {
		if errors.Is(err, storage.ErrAssetNotFound) {
			return Forecast{}, ErrAssetNotFound
		}
		return Forecast{}, err
	}

	eventRecord, newsRecord, err := s.resolveEventContext(ctx, asset.ID, req.EventID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEventNotFound):
			return Forecast{}, ErrEventNotFound
		case errors.Is(err, ErrEventAssetMismatch):
			return Forecast{}, ErrEventAssetMismatch
		default:
			return Forecast{}, err
		}
	}

	assetIndicator, err := s.loadLatestIndicator(ctx, asset.ID)
	if err != nil {
		return Forecast{}, err
	}

	marketContext, err := s.buildMarketContext(ctx)
	if err != nil {
		return Forecast{}, err
	}

	aiInput := ai.Input{
		Horizon: normalizeHorizon(req.Horizon),
		Asset: ai.AssetContext{
			ID:        asset.ID,
			Ticker:    asset.Ticker,
			Name:      asset.Name,
			AssetType: asset.AssetType,
			Sector:    asset.Sector,
			Currency:  asset.Currency,
		},
		Indicators: mapIndicatorContext(assetIndicator),
		Market: ai.MarketContext{
			Label:       marketContext.Label,
			Score:       marketContext.Score,
			Explanation: marketContext.Explanation,
		},
	}

	var eventID *string
	if eventRecord != nil {
		eventID = stringPointer(eventRecord.ID)
		aiInput.Event = &ai.EventContext{
			ID:      eventRecord.ID,
			Type:    eventRecord.EventType,
			Summary: eventRecord.Summary,
		}
	}

	if newsRecord != nil {
		aiInput.News = &ai.NewsContext{
			ID:          newsRecord.ID,
			Title:       newsRecord.Title,
			Summary:     newsRecord.Summary,
			Body:        newsRecord.Body,
			SourceName:  newsRecord.SourceName,
			PublishedAt: newsRecord.PublishedAt,
		}
	}

	aiOutput, err := s.aiClient.Generate(ctx, aiInput)
	if err != nil {
		return Forecast{}, err
	}

	keyFactorsJSON, err := json.Marshal(aiOutput.KeyFactors)
	if err != nil {
		return Forecast{}, fmt.Errorf("marshal key factors: %w", err)
	}

	var preparedRequestJSON json.RawMessage
	if aiOutput.PreparedRequest != nil {
		preparedRequestJSON, err = json.Marshal(aiOutput.PreparedRequest)
		if err != nil {
			return Forecast{}, fmt.Errorf("marshal prepared ai request: %w", err)
		}
	}

	record, err := s.forecastsRepository.Create(ctx, storage.CreateForecastParams{
		AssetID:                  asset.ID,
		EventID:                  eventID,
		ForecastHorizon:          aiInput.Horizon,
		ForecastTime:             time.Now().UTC(),
		DirectionLabel:           aiOutput.Direction,
		SignalStrength:           aiOutput.Strength,
		ConfidenceScore:          aiOutput.Confidence,
		Explanation:              aiOutput.Explanation,
		AIMode:                   aiOutput.Mode,
		ModelName:                aiOutput.Model,
		MarketContextLabel:       marketContext.Label,
		MarketContextScore:       marketContext.Score,
		MarketContextExplanation: marketContext.Explanation,
		KeyFactorsJSON:           keyFactorsJSON,
		PreparedRequestJSON:      preparedRequestJSON,
	})
	if err != nil {
		return Forecast{}, err
	}

	return mapForecast(record)
}

func (s *Service) Latest(ctx context.Context) (Forecast, error) {
	record, err := s.forecastsRepository.GetLatest(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrForecastNotFound) {
			return Forecast{}, ErrForecastNotFound
		}
		return Forecast{}, err
	}

	return mapForecast(record)
}

func (s *Service) resolveEventContext(ctx context.Context, assetID, requestedEventID string) (*storage.EventRecord, *storage.NewsItemRecord, error) {
	var (
		event storage.EventRecord
		err   error
	)

	if strings.TrimSpace(requestedEventID) != "" {
		event, err = s.eventsRepository.GetByID(ctx, requestedEventID)
		if err != nil {
			return nil, nil, err
		}
		if event.AssetID.Valid && event.AssetID.String != assetID {
			return nil, nil, ErrEventAssetMismatch
		}
	} else {
		event, err = s.eventsRepository.GetLatestRelevant(ctx, assetID)
		if err != nil {
			if errors.Is(err, storage.ErrEventNotFound) {
				return nil, nil, nil
			}
			return nil, nil, err
		}
	}

	newsItem, err := s.newsRepository.GetByID(ctx, event.NewsItemID)
	if err != nil {
		return nil, nil, err
	}

	return &event, &newsItem, nil
}

func (s *Service) loadLatestIndicator(ctx context.Context, assetID string) (*storage.TechnicalIndicatorRecord, error) {
	item, err := s.indicatorsRepository.GetLatestByAsset(ctx, assetID, dailyTimeframe)
	if err != nil {
		if errors.Is(err, storage.ErrTechnicalIndicatorNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

func (s *Service) buildMarketContext(ctx context.Context) (derivedMarketContext, error) {
	marketAsset, err := s.assetsRepository.GetByTicker(ctx, marketContextAssetTicker)
	if err != nil {
		if errors.Is(err, storage.ErrAssetNotFound) {
			return derivedMarketContext{
				Label:       "stable",
				Score:       0.20,
				Explanation: "Market context falls back to a stable baseline because IMOEX is missing from the asset catalog.",
			}, nil
		}
		return derivedMarketContext{}, err
	}

	indicator, err := s.loadLatestIndicator(ctx, marketAsset.ID)
	if err != nil {
		return derivedMarketContext{}, err
	}

	if indicator == nil {
		return derivedMarketContext{
			Label:       "stable",
			Score:       0.20,
			Explanation: "Market context falls back to a stable baseline because latest IMOEX indicators are not available yet.",
		}, nil
	}

	score := 0.20
	reasons := make([]string, 0)

	if indicator.WeeklyReturn.Valid {
		switch {
		case indicator.WeeklyReturn.Float64 <= -0.03:
			score += 0.30
			reasons = append(reasons, "negative weekly return in IMOEX")
		case indicator.WeeklyReturn.Float64 >= 0.03:
			score -= 0.05
			reasons = append(reasons, "positive weekly return in IMOEX")
		}
	}

	if indicator.Volatility.Valid {
		switch {
		case indicator.Volatility.Float64 >= 0.45:
			score += 0.25
			reasons = append(reasons, "elevated market volatility")
		case indicator.Volatility.Float64 >= 0.30:
			score += 0.10
			reasons = append(reasons, "moderately high market volatility")
		}
	}

	if indicator.TrendDirection.Valid {
		switch indicator.TrendDirection.String {
		case "down":
			score += 0.20
			reasons = append(reasons, "downward market trend")
		case "flat":
			score += 0.05
			reasons = append(reasons, "flat market trend")
		}
	}

	if indicator.ChannelPosition.Valid && indicator.ChannelPosition.Float64 < 0.25 {
		score += 0.10
		reasons = append(reasons, "price sits near the lower edge of the local range")
	}

	if indicator.RSI.Valid {
		switch {
		case indicator.RSI.Float64 < 35:
			score += 0.10
			reasons = append(reasons, "weak momentum by RSI")
		case indicator.RSI.Float64 > 65:
			score += 0.05
			reasons = append(reasons, "overheated momentum by RSI")
		}
	}

	score = clamp(score, 0, 1)
	label := deriveMarketLabel(score)

	explanation := "Market context is based on latest IMOEX technical indicators."
	if len(reasons) > 0 {
		explanation = fmt.Sprintf("Market context is based on latest IMOEX indicators: %s.", strings.Join(reasons, ", "))
	}

	return derivedMarketContext{
		Label:       label,
		Score:       round2(score),
		Explanation: explanation,
	}, nil
}

func mapIndicatorContext(item *storage.TechnicalIndicatorRecord) ai.IndicatorContext {
	if item == nil {
		return ai.IndicatorContext{
			Timeframe:         dailyTimeframe,
			CalculationStatus: "insufficient_data",
		}
	}

	return ai.IndicatorContext{
		Timeframe:         item.Timeframe,
		IndicatorTime:     item.IndicatorTime,
		WeeklyReturn:      nullFloat64ToPointer(item.WeeklyReturn),
		RSI:               nullFloat64ToPointer(item.RSI),
		Volatility:        nullFloat64ToPointer(item.Volatility),
		TrendDirection:    nullStringToPointer(item.TrendDirection),
		ChannelPosition:   nullFloat64ToPointer(item.ChannelPosition),
		CalculationStatus: item.CalculationStatus,
	}
}

func mapForecast(record storage.ForecastRecord) (Forecast, error) {
	keyFactors := make([]string, 0)
	if len(record.KeyFactorsJSON) > 0 {
		if err := json.Unmarshal(record.KeyFactorsJSON, &keyFactors); err != nil {
			return Forecast{}, fmt.Errorf("unmarshal forecast key factors: %w", err)
		}
	}

	return Forecast{
		ID:           record.ID,
		AssetID:      record.AssetID,
		AssetTicker:  record.AssetTicker,
		AssetName:    record.AssetName,
		EventID:      nullStringToPointer(record.EventID),
		EventType:    nullStringToPointer(record.EventType),
		EventSummary: nullStringToPointer(record.EventSummary),
		Horizon:      record.ForecastHorizon,
		Direction:    record.DirectionLabel,
		Strength:     record.SignalStrength,
		Confidence:   record.ConfidenceScore,
		Explanation:  record.Explanation,
		GeneratedAt:  record.ForecastTime,
		AIMode:       record.AIMode,
		Model:        record.ModelName,
		MarketContext: MarketContext{
			Label:       record.MarketContextLabel,
			Score:       record.MarketContextScore,
			Explanation: record.MarketContextExplanation,
		},
		KeyFactors:      keyFactors,
		PreparedRequest: cloneRawJSON(record.PreparedRequestJSON),
	}, nil
}

func normalizeHorizon(value string) string {
	if strings.TrimSpace(value) == "" {
		return defaultForecastHorizon
	}
	return strings.TrimSpace(value)
}

func deriveMarketLabel(score float64) string {
	switch {
	case score < 0.25:
		return "stable"
	case score < 0.50:
		return "cautious"
	case score < 0.75:
		return "stressed"
	default:
		return "crisis"
	}
}

func clamp(value, minValue, maxValue float64) float64 {
	switch {
	case value < minValue:
		return minValue
	case value > maxValue:
		return maxValue
	default:
		return value
	}
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func stringPointer(value string) *string {
	return &value
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

func cloneRawJSON(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return nil
	}

	result := make(json.RawMessage, len(value))
	copy(result, value)
	return result
}
