package regime

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"diploma-market-ai/02_product/backend/internal/storage"
)

const (
	calculationModelRuleBased = "rule_based_mvp"
	dailyTimeframe            = "1d"
	marketTicker              = "IMOEX"
	recentEventLimit          = 20
)

var (
	breadthTickers   = []string{"SBER", "LKOH", "GAZP", "YDEX"}
	commodityTickers = []string{"BRENT", "NATGAS"}
)

type Service struct {
	assetsRepository     *storage.AssetsRepository
	eventsRepository     *storage.EventsRepository
	indicatorsRepository *storage.TechnicalIndicatorsRepository
	regimesRepository    *storage.MarketRegimesRepository
}

type MarketRegime struct {
	ID               string          `json:"id"`
	RegimeScore      float64         `json:"regime_score"`
	RegimeLabel      string          `json:"regime_label"`
	SubScores        RegimeSubScores `json:"sub_scores"`
	Summary          string          `json:"summary"`
	Explanation      string          `json:"explanation"`
	CalculationModel string          `json:"calculation_model"`
	CalculatedAt     time.Time       `json:"calculated_at"`
}

type RegimeSubScores struct {
	MarketStress    float64 `json:"market_stress"`
	NewsStress      float64 `json:"news_stress"`
	MacroStress     float64 `json:"macro_stress"`
	CommodityStress float64 `json:"commodity_stress"`
	BreadthStress   float64 `json:"breadth_stress"`
}

type componentScore struct {
	Name       string
	Title      string
	Score      float64
	Reasons    []string
	ObservedAt time.Time
}

type calculatedRegime struct {
	RegimeTime  time.Time
	RegimeScore float64
	RegimeLabel string
	SubScores   RegimeSubScores
	Summary     string
	Explanation string
}

func NewService(store *storage.Postgres) *Service {
	return &Service{
		assetsRepository:     storage.NewAssetsRepository(store),
		eventsRepository:     storage.NewEventsRepository(store),
		indicatorsRepository: storage.NewTechnicalIndicatorsRepository(store),
		regimesRepository:    storage.NewMarketRegimesRepository(store),
	}
}

func (s *Service) Current(ctx context.Context) (MarketRegime, error) {
	calculated, err := s.calculateCurrent(ctx)
	if err != nil {
		return MarketRegime{}, err
	}

	record, err := s.regimesRepository.Save(ctx, storage.SaveMarketRegimeParams{
		RegimeTime:           calculated.RegimeTime,
		RegimeLabel:          calculated.RegimeLabel,
		RegimeScore:          calculated.RegimeScore,
		MarketStressScore:    calculated.SubScores.MarketStress,
		NewsStressScore:      calculated.SubScores.NewsStress,
		MacroStressScore:     calculated.SubScores.MacroStress,
		CommodityStressScore: calculated.SubScores.CommodityStress,
		BreadthStressScore:   calculated.SubScores.BreadthStress,
		Summary:              calculated.Summary,
		Explanation:          calculated.Explanation,
		CalculationModel:     calculationModelRuleBased,
	})
	if err != nil {
		return MarketRegime{}, err
	}

	return mapMarketRegime(record), nil
}

func (s *Service) calculateCurrent(ctx context.Context) (calculatedRegime, error) {
	recentEvents, err := s.eventsRepository.ListRecent(ctx, recentEventLimit)
	if err != nil {
		return calculatedRegime{}, err
	}

	marketComponent, err := s.calculateMarketStress(ctx)
	if err != nil {
		return calculatedRegime{}, err
	}

	newsComponent := s.calculateNewsStress(recentEvents)
	macroComponent := s.calculateMacroStress(recentEvents)

	commodityComponent, err := s.calculateCommodityStress(ctx, recentEvents)
	if err != nil {
		return calculatedRegime{}, err
	}

	breadthComponent, err := s.calculateBreadthStress(ctx)
	if err != nil {
		return calculatedRegime{}, err
	}

	components := []componentScore{
		marketComponent,
		newsComponent,
		macroComponent,
		commodityComponent,
		breadthComponent,
	}

	regimeScore := round2(
		marketComponent.Score*0.30 +
			newsComponent.Score*0.20 +
			macroComponent.Score*0.20 +
			commodityComponent.Score*0.15 +
			breadthComponent.Score*0.15,
	)

	regimeLabel := deriveRegimeLabel(regimeScore)
	summary := buildSummary(regimeLabel, components)
	explanation := buildExplanation(regimeLabel, components)
	regimeTime := maxTime(
		marketComponent.ObservedAt,
		newsComponent.ObservedAt,
		macroComponent.ObservedAt,
		commodityComponent.ObservedAt,
		breadthComponent.ObservedAt,
	)
	if regimeTime.IsZero() {
		regimeTime = time.Now().UTC().Truncate(time.Minute)
	}

	return calculatedRegime{
		RegimeTime:  regimeTime,
		RegimeScore: regimeScore,
		RegimeLabel: regimeLabel,
		SubScores: RegimeSubScores{
			MarketStress:    marketComponent.Score,
			NewsStress:      newsComponent.Score,
			MacroStress:     macroComponent.Score,
			CommodityStress: commodityComponent.Score,
			BreadthStress:   breadthComponent.Score,
		},
		Summary:     summary,
		Explanation: explanation,
	}, nil
}

func (s *Service) calculateMarketStress(ctx context.Context) (componentScore, error) {
	indicator, err := s.loadLatestIndicatorByTicker(ctx, marketTicker)
	if err != nil {
		return componentScore{}, err
	}

	if indicator == nil {
		return componentScore{
			Name:    "market_stress",
			Title:   "market stress",
			Score:   0.35,
			Reasons: []string{"latest IMOEX indicators are not available, so the block stays close to neutral"},
		}, nil
	}

	score := 0.20
	reasons := make([]string, 0)

	if indicator.WeeklyReturn.Valid {
		switch {
		case indicator.WeeklyReturn.Float64 <= -0.07:
			score += 0.35
			reasons = append(reasons, "IMOEX weekly return is sharply negative")
		case indicator.WeeklyReturn.Float64 <= -0.03:
			score += 0.20
			reasons = append(reasons, "IMOEX weekly return remains negative")
		case indicator.WeeklyReturn.Float64 >= 0.05:
			score -= 0.05
			reasons = append(reasons, "IMOEX weekly return stays positive")
		}
	}

	if indicator.Volatility.Valid {
		switch {
		case indicator.Volatility.Float64 >= 0.55:
			score += 0.25
			reasons = append(reasons, "market volatility is elevated")
		case indicator.Volatility.Float64 >= 0.35:
			score += 0.15
			reasons = append(reasons, "market volatility is above normal")
		case indicator.Volatility.Float64 <= 0.18:
			score -= 0.05
			reasons = append(reasons, "market volatility remains contained")
		}
	}

	if indicator.TrendDirection.Valid {
		switch indicator.TrendDirection.String {
		case "down":
			score += 0.15
			reasons = append(reasons, "short-term trend points down")
		case "flat":
			score += 0.05
			reasons = append(reasons, "market trend is flat rather than supportive")
		case "up":
			score -= 0.05
			reasons = append(reasons, "trend still points up")
		}
	}

	if indicator.ChannelPosition.Valid {
		switch {
		case indicator.ChannelPosition.Float64 <= 0.20:
			score += 0.10
			reasons = append(reasons, "IMOEX closes near the lower edge of its local range")
		case indicator.ChannelPosition.Float64 >= 0.80:
			score -= 0.05
			reasons = append(reasons, "IMOEX holds near the upper edge of its range")
		}
	}

	if indicator.RSI.Valid {
		switch {
		case indicator.RSI.Float64 < 30:
			score += 0.10
			reasons = append(reasons, "RSI shows weak momentum")
		case indicator.RSI.Float64 > 70:
			score += 0.05
			reasons = append(reasons, "RSI shows overheated conditions")
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "IMOEX technical context remains close to neutral")
	}

	return componentScore{
		Name:       "market_stress",
		Title:      "market stress",
		Score:      round2(clamp(score, 0, 1)),
		Reasons:    reasons,
		ObservedAt: indicator.IndicatorTime,
	}, nil
}

func (s *Service) calculateNewsStress(events []storage.EventRecord) componentScore {
	recent := filterEventsByAge(events, 14*24*time.Hour)
	if len(recent) == 0 {
		return componentScore{
			Name:    "news_stress",
			Title:   "news stress",
			Score:   0.30,
			Reasons: []string{"no dense cluster of recent events was detected, so the news block stays neutral"},
		}
	}

	var (
		total       float64
		totalWeight float64
		negative    int
		systemic    int
		reasons     []string
	)

	for _, event := range recent {
		text := strings.TrimSpace(strings.Join([]string{event.NewsTitle, event.Summary}, " "))
		score := clamp(
			newsStressFromEventType(event.EventType)+textStressAdjustment(text),
			0,
			1,
		)
		weight := recencyWeight(eventPublishedAt(event))
		total += score * weight
		totalWeight += weight

		if score >= 0.60 {
			negative++
		}
		if !event.AssetID.Valid {
			systemic++
		}
		if score >= 0.55 && len(reasons) < 3 {
			reasons = append(reasons, describeEventForReason(event))
		}
	}

	score := 0.30
	if totalWeight > 0 {
		score = total / totalWeight
	}
	if negative >= 3 {
		score += 0.05
		reasons = append(reasons, "several recent events point in a negative direction at once")
	}
	if systemic >= 2 {
		score += 0.05
		reasons = append(reasons, "part of the recent flow affects the market broadly rather than a single asset")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "recent event flow remains mixed without a clear stress cluster")
	}

	return componentScore{
		Name:       "news_stress",
		Title:      "news stress",
		Score:      round2(clamp(score, 0, 1)),
		Reasons:    uniqueStrings(reasons),
		ObservedAt: eventPublishedAt(recent[0]),
	}
}

func (s *Service) calculateMacroStress(events []storage.EventRecord) componentScore {
	macroEvents := make([]storage.EventRecord, 0)
	for _, event := range filterEventsByAge(events, 30*24*time.Hour) {
		switch event.EventType {
		case "key_rate_hike", "key_rate_hold", "key_rate_cut", "monetary_policy":
			macroEvents = append(macroEvents, event)
		}
	}

	if len(macroEvents) == 0 {
		return componentScore{
			Name:    "macro_stress",
			Title:   "macro stress",
			Score:   0.35,
			Reasons: []string{"there is no fresh macro trigger in the current window, so the macro block stays neutral"},
		}
	}

	var (
		total       float64
		totalWeight float64
		reasons     []string
	)

	for _, event := range macroEvents {
		text := strings.TrimSpace(strings.Join([]string{event.NewsTitle, event.Summary}, " "))
		score := clamp(
			macroStressFromEventType(event.EventType)+textStressAdjustment(text),
			0,
			1,
		)
		weight := recencyWeight(eventPublishedAt(event))
		total += score * weight
		totalWeight += weight
		if len(reasons) < 2 {
			reasons = append(reasons, describeEventForReason(event))
		}
	}

	score := 0.35
	if totalWeight > 0 {
		score = total / totalWeight
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "macro signals remain mixed")
	}

	return componentScore{
		Name:       "macro_stress",
		Title:      "macro stress",
		Score:      round2(clamp(score, 0, 1)),
		Reasons:    uniqueStrings(reasons),
		ObservedAt: eventPublishedAt(macroEvents[0]),
	}
}

func (s *Service) calculateCommodityStress(ctx context.Context, events []storage.EventRecord) (componentScore, error) {
	scores := make([]float64, 0, len(commodityTickers))
	reasons := make([]string, 0)
	observedAt := time.Time{}

	for _, ticker := range commodityTickers {
		indicator, err := s.loadLatestIndicatorByTicker(ctx, ticker)
		if err != nil {
			return componentScore{}, err
		}
		if indicator == nil {
			continue
		}

		score, itemReasons := commodityStressFromIndicator(ticker, *indicator)
		scores = append(scores, score)
		reasons = append(reasons, itemReasons...)
		observedAt = maxTime(observedAt, indicator.IndicatorTime)
	}

	if len(scores) > 0 {
		return componentScore{
			Name:       "commodity_stress",
			Title:      "commodity stress",
			Score:      round2(clamp(average(scores), 0, 1)),
			Reasons:    uniqueStrings(reasons),
			ObservedAt: observedAt,
		}, nil
	}

	commodityEvents := make([]storage.EventRecord, 0)
	for _, event := range filterEventsByAge(events, 14*24*time.Hour) {
		if event.EventType == "commodity_oil" || event.EventType == "commodity_gas" {
			commodityEvents = append(commodityEvents, event)
		}
	}

	if len(commodityEvents) == 0 {
		return componentScore{
			Name:    "commodity_stress",
			Title:   "commodity stress",
			Score:   0.30,
			Reasons: []string{"direct commodity market inputs are not loaded yet, so the block stays neutral"},
		}, nil
	}

	var (
		total       float64
		totalWeight float64
	)
	for _, event := range commodityEvents {
		text := strings.TrimSpace(strings.Join([]string{event.NewsTitle, event.Summary}, " "))
		score := clamp(0.50+textStressAdjustment(text), 0, 1)
		weight := recencyWeight(eventPublishedAt(event))
		total += score * weight
		totalWeight += weight
		if len(reasons) < 2 {
			reasons = append(reasons, describeEventForReason(event))
		}
	}

	score := 0.30
	if totalWeight > 0 {
		score = total / totalWeight
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "commodity news flow remains mixed")
	}

	return componentScore{
		Name:       "commodity_stress",
		Title:      "commodity stress",
		Score:      round2(clamp(score, 0, 1)),
		Reasons:    uniqueStrings(reasons),
		ObservedAt: eventPublishedAt(commodityEvents[0]),
	}, nil
}

func (s *Service) calculateBreadthStress(ctx context.Context) (componentScore, error) {
	var (
		available  int
		negative   int
		downTrend  int
		weakRange  int
		reasons    []string
		observedAt time.Time
	)

	for _, ticker := range breadthTickers {
		indicator, err := s.loadLatestIndicatorByTicker(ctx, ticker)
		if err != nil {
			return componentScore{}, err
		}
		if indicator == nil {
			continue
		}

		available++
		observedAt = maxTime(observedAt, indicator.IndicatorTime)

		if indicator.WeeklyReturn.Valid && indicator.WeeklyReturn.Float64 < 0 {
			negative++
		}
		if indicator.TrendDirection.Valid && indicator.TrendDirection.String == "down" {
			downTrend++
		}
		if indicator.ChannelPosition.Valid && indicator.ChannelPosition.Float64 < 0.35 {
			weakRange++
		}
	}

	if available == 0 {
		return componentScore{
			Name:    "breadth_stress",
			Title:   "breadth stress",
			Score:   0.35,
			Reasons: []string{"key asset indicators are not available yet, so breadth is kept near neutral"},
		}, nil
	}

	negativeShare := float64(negative) / float64(available)
	downShare := float64(downTrend) / float64(available)
	weakShare := float64(weakRange) / float64(available)

	score := 0.20 + negativeShare*0.40 + downShare*0.25 + weakShare*0.15

	switch {
	case negativeShare >= 0.75:
		reasons = append(reasons, "most key assets show negative weekly returns")
	case negativeShare >= 0.50:
		reasons = append(reasons, "a broad part of the key asset set is trading in negative territory")
	default:
		reasons = append(reasons, "breadth remains mixed rather than broadly weak")
	}

	if downShare >= 0.75 {
		reasons = append(reasons, "downtrends dominate across key assets")
	}
	if weakShare >= 0.50 {
		reasons = append(reasons, "many key assets stay near the lower part of their local ranges")
	}

	return componentScore{
		Name:       "breadth_stress",
		Title:      "breadth stress",
		Score:      round2(clamp(score, 0, 1)),
		Reasons:    reasons,
		ObservedAt: observedAt,
	}, nil
}

func (s *Service) loadLatestIndicatorByTicker(ctx context.Context, ticker string) (*storage.TechnicalIndicatorRecord, error) {
	asset, err := s.assetsRepository.GetByTicker(ctx, ticker)
	if err != nil {
		if errors.Is(err, storage.ErrAssetNotFound) {
			return nil, nil
		}
		return nil, err
	}

	indicator, err := s.indicatorsRepository.GetLatestByAsset(ctx, asset.ID, dailyTimeframe)
	if err != nil {
		if errors.Is(err, storage.ErrTechnicalIndicatorNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &indicator, nil
}

func mapMarketRegime(record storage.MarketRegimeRecord) MarketRegime {
	return MarketRegime{
		ID:          record.ID,
		RegimeScore: record.RegimeScore,
		RegimeLabel: record.RegimeLabel,
		SubScores: RegimeSubScores{
			MarketStress:    record.MarketStressScore,
			NewsStress:      record.NewsStressScore,
			MacroStress:     record.MacroStressScore,
			CommodityStress: record.CommodityStressScore,
			BreadthStress:   record.BreadthStressScore,
		},
		Summary:          record.Summary,
		Explanation:      record.Explanation,
		CalculationModel: record.CalculationModel,
		CalculatedAt:     record.RegimeTime,
	}
}

func buildSummary(regimeLabel string, components []componentScore) string {
	dominant := dominantComponents(components)
	if len(dominant) == 0 {
		return fmt.Sprintf(
			"Rule-based MVP crisisometer sets the market to %s because no single stress block dominates the current inputs.",
			regimeLabel,
		)
	}

	return fmt.Sprintf(
		"Rule-based MVP crisisometer sets the market to %s. Main contributors: %s.",
		regimeLabel,
		strings.Join(dominant, ", "),
	)
}

func buildExplanation(regimeLabel string, components []componentScore) string {
	parts := []string{
		"This is a temporary rule-based MVP implementation of the crisisometer. It combines market, news, macro, commodity and breadth blocks using transparent weights.",
		fmt.Sprintf("Current overall label: %s.", regimeLabel),
	}

	for _, component := range sortedComponentsDesc(components) {
		if len(component.Reasons) == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s.", component.Title, strings.Join(component.Reasons, "; ")))
	}

	return strings.Join(parts, " ")
}

func dominantComponents(components []componentScore) []string {
	sorted := sortedComponentsDesc(components)
	result := make([]string, 0, 2)
	for _, component := range sorted {
		if component.Score < 0.50 {
			continue
		}
		result = append(result, fmt.Sprintf("%s %.2f", component.Title, component.Score))
		if len(result) == 2 {
			break
		}
	}
	return result
}

func sortedComponentsDesc(components []componentScore) []componentScore {
	cloned := make([]componentScore, len(components))
	copy(cloned, components)

	sort.Slice(cloned, func(i, j int) bool {
		if cloned[i].Score == cloned[j].Score {
			return cloned[i].Name < cloned[j].Name
		}
		return cloned[i].Score > cloned[j].Score
	})

	return cloned
}

func deriveRegimeLabel(score float64) string {
	switch {
	case score < 0.20:
		return "stable"
	case score < 0.40:
		return "moderate_tension"
	case score < 0.60:
		return "elevated_stress"
	case score < 0.80:
		return "pre_crisis"
	default:
		return "crisis"
	}
}

func filterEventsByAge(events []storage.EventRecord, maxAge time.Duration) []storage.EventRecord {
	if len(events) == 0 {
		return nil
	}

	now := time.Now().UTC()
	result := make([]storage.EventRecord, 0, len(events))
	for _, event := range events {
		publishedAt := eventPublishedAt(event)
		if publishedAt.IsZero() || now.Sub(publishedAt) <= maxAge {
			result = append(result, event)
		}
	}

	return result
}

func eventPublishedAt(event storage.EventRecord) time.Time {
	if !event.PublishedAt.IsZero() {
		return event.PublishedAt
	}
	return event.ExtractedAt
}

func recencyWeight(observedAt time.Time) float64 {
	if observedAt.IsZero() {
		return 0.50
	}

	age := time.Since(observedAt)
	switch {
	case age <= 72*time.Hour:
		return 1.00
	case age <= 7*24*time.Hour:
		return 0.80
	case age <= 14*24*time.Hour:
		return 0.60
	case age <= 30*24*time.Hour:
		return 0.40
	default:
		return 0.20
	}
}

func newsStressFromEventType(eventType string) float64 {
	switch eventType {
	case "sanctions":
		return 0.90
	case "key_rate_hike":
		return 0.75
	case "monetary_policy":
		return 0.60
	case "key_rate_hold":
		return 0.45
	case "commodity_oil", "commodity_gas":
		return 0.50
	case "financial_results":
		return 0.30
	case "dividend":
		return 0.20
	case "key_rate_cut":
		return 0.25
	default:
		return 0.35
	}
}

func macroStressFromEventType(eventType string) float64 {
	switch eventType {
	case "key_rate_hike":
		return 0.85
	case "monetary_policy":
		return 0.60
	case "key_rate_hold":
		return 0.55
	case "key_rate_cut":
		return 0.25
	default:
		return 0.35
	}
}

func commodityStressFromIndicator(ticker string, indicator storage.TechnicalIndicatorRecord) (float64, []string) {
	score := 0.30
	reasons := make([]string, 0)

	if indicator.WeeklyReturn.Valid {
		switch {
		case indicator.WeeklyReturn.Float64 <= -0.10:
			score += 0.30
			reasons = append(reasons, fmt.Sprintf("%s weekly return is sharply negative", ticker))
		case indicator.WeeklyReturn.Float64 <= -0.05:
			score += 0.18
			reasons = append(reasons, fmt.Sprintf("%s weekly return remains under pressure", ticker))
		case indicator.WeeklyReturn.Float64 >= 0.08:
			score -= 0.10
			reasons = append(reasons, fmt.Sprintf("%s weekly return is positive", ticker))
		}
	}

	if indicator.Volatility.Valid {
		switch {
		case indicator.Volatility.Float64 >= 0.60:
			score += 0.15
			reasons = append(reasons, fmt.Sprintf("%s volatility is high", ticker))
		case indicator.Volatility.Float64 >= 0.35:
			score += 0.05
			reasons = append(reasons, fmt.Sprintf("%s volatility is elevated", ticker))
		}
	}

	if indicator.TrendDirection.Valid && indicator.TrendDirection.String == "down" {
		score += 0.12
		reasons = append(reasons, fmt.Sprintf("%s trend points down", ticker))
	}

	if indicator.ChannelPosition.Valid && indicator.ChannelPosition.Float64 < 0.30 {
		score += 0.08
		reasons = append(reasons, fmt.Sprintf("%s trades near the lower edge of its range", ticker))
	}

	if indicator.RSI.Valid && indicator.RSI.Float64 < 35 {
		score += 0.05
		reasons = append(reasons, fmt.Sprintf("%s momentum is weak", ticker))
	}

	if len(reasons) == 0 {
		reasons = append(reasons, fmt.Sprintf("%s commodity context remains neutral", ticker))
	}

	return round2(clamp(score, 0, 1)), reasons
}

func describeEventForReason(event storage.EventRecord) string {
	switch event.EventType {
	case "sanctions":
		return "sanction-related news stays in the recent flow"
	case "key_rate_hike":
		return "recent macro event points to tighter monetary conditions"
	case "key_rate_hold":
		return "recent macro event keeps the rate at a restrictive level"
	case "key_rate_cut":
		return "recent macro event signals monetary easing"
	case "commodity_oil":
		return "recent oil-related event affects the external backdrop"
	case "commodity_gas":
		return "recent gas-related event affects the external backdrop"
	case "financial_results":
		return "recent corporate results influence the tone of the flow"
	case "dividend":
		return "recent dividend-related event affects sentiment"
	default:
		return "recent event flow remains active"
	}
}

func textStressAdjustment(text string) float64 {
	normalized := strings.ToUpper(text)

	negativeKeywords := []string{
		"SANCTION",
		"INFLATION",
		"PRESSURE",
		"SLOWDOWN",
		"DECLINE",
		"DROP",
		"FALL",
		"CRISIS",
		"RISK",
		"VOLATILITY",
		"CONFLICT",
		"TIGHT",
		"WEAK",
		"LOSS",
	}
	positiveKeywords := []string{
		"GROWTH",
		"SUPPORT",
		"RECOVERY",
		"DIVIDEND",
		"APPROVED",
		"CUT",
		"EASING",
		"IMPROV",
		"STRONG",
	}

	var adjustment float64
	for _, keyword := range negativeKeywords {
		if strings.Contains(normalized, keyword) {
			adjustment += 0.08
		}
	}
	for _, keyword := range positiveKeywords {
		if strings.Contains(normalized, keyword) {
			adjustment -= 0.06
		}
	}

	return clamp(adjustment, -0.18, 0.25)
}

func maxTime(values ...time.Time) time.Time {
	result := time.Time{}
	for _, value := range values {
		if value.After(result) {
			result = value
		}
	}
	return result
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var total float64
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
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
