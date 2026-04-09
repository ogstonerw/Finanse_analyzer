package ai

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	ModeFallback = "fallback"
	ModePrepare  = "prepare"
)

type Config struct {
	Mode     string
	Provider string
	Model    string
	Endpoint string
	APIKey   string
}

type Client struct {
	config Config
}

type Input struct {
	Horizon    string           `json:"horizon"`
	Asset      AssetContext     `json:"asset"`
	Event      *EventContext    `json:"event,omitempty"`
	News       *NewsContext     `json:"news,omitempty"`
	Indicators IndicatorContext `json:"indicators"`
	Market     MarketContext    `json:"market"`
}

type AssetContext struct {
	ID        string `json:"id"`
	Ticker    string `json:"ticker"`
	Name      string `json:"name"`
	AssetType string `json:"asset_type"`
	Sector    string `json:"sector"`
	Currency  string `json:"currency"`
}

type EventContext struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

type NewsContext struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Body        string    `json:"body"`
	SourceName  string    `json:"source_name"`
	PublishedAt time.Time `json:"published_at"`
}

type IndicatorContext struct {
	Timeframe         string    `json:"timeframe"`
	IndicatorTime     time.Time `json:"indicator_time"`
	WeeklyReturn      *float64  `json:"weekly_return,omitempty"`
	RSI               *float64  `json:"rsi,omitempty"`
	Volatility        *float64  `json:"volatility,omitempty"`
	TrendDirection    *string   `json:"trend_direction,omitempty"`
	ChannelPosition   *float64  `json:"channel_position,omitempty"`
	CalculationStatus string    `json:"calculation_status"`
}

type MarketContext struct {
	Label       string  `json:"label"`
	Score       float64 `json:"score"`
	Explanation string  `json:"explanation"`
}

type PreparedRequest struct {
	Provider string         `json:"provider"`
	Endpoint string         `json:"endpoint"`
	Model    string         `json:"model"`
	Payload  map[string]any `json:"payload"`
}

type Output struct {
	Direction       string           `json:"direction"`
	Strength        float64          `json:"strength"`
	Confidence      float64          `json:"confidence"`
	Explanation     string           `json:"explanation"`
	KeyFactors      []string         `json:"key_factors"`
	Mode            string           `json:"mode"`
	Model           string           `json:"model"`
	PreparedRequest *PreparedRequest `json:"prepared_request,omitempty"`
}

type scoredFactor struct {
	label        string
	contribution float64
}

func NewClient(cfg Config) *Client {
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode != ModePrepare {
		mode = ModeFallback
	}

	provider := strings.TrimSpace(cfg.Provider)
	if provider == "" {
		provider = "openai"
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		if mode == ModePrepare {
			model = "future-openai-model"
		} else {
			model = "fallback-rule-engine"
		}
	}

	return &Client{
		config: Config{
			Mode:     mode,
			Provider: provider,
			Model:    model,
			Endpoint: strings.TrimSpace(cfg.Endpoint),
			APIKey:   strings.TrimSpace(cfg.APIKey),
		},
	}
}

func (c *Client) Generate(_ context.Context, input Input) (Output, error) {
	output := generateFallback(input)
	output.Mode = c.config.Mode
	output.Model = c.config.Model

	if c.config.Mode == ModePrepare {
		prepared := c.BuildPreparedRequest(input)
		output.PreparedRequest = &prepared
	}

	return output, nil
}

func (c *Client) BuildPreparedRequest(input Input) PreparedRequest {
	payload := map[string]any{
		"instructions": "Analyze the structured market context and return direction, strength, confidence, explanation, and key factors for a one-week market reaction forecast.",
		"expected_output": map[string]any{
			"direction":   "up | neutral | down",
			"strength":    "normalized float in range 0..1",
			"confidence":  "normalized float in range 0..1",
			"explanation": "short human-readable explanation",
			"key_factors": []string{"factor"},
		},
		"input": input,
	}

	return PreparedRequest{
		Provider: c.config.Provider,
		Endpoint: c.config.Endpoint,
		Model:    c.config.Model,
		Payload:  payload,
	}
}

func generateFallback(input Input) Output {
	eventScore, eventFactors := scoreEvent(input)
	technicalScore, technicalFactors := scoreTechnical(input)
	marketScore, marketFactors := scoreMarket(input)

	combined := clamp(eventScore+technicalScore+marketScore, -1, 1)
	direction := deriveDirection(combined)
	strength := round2(math.Abs(combined))
	confidence := round2(calculateConfidence(input, direction, eventScore, technicalScore))

	allFactors := append(append(eventFactors, technicalFactors...), marketFactors...)
	keyFactors := pickTopFactors(allFactors, 4)
	explanation := buildExplanation(input, direction, keyFactors)

	return Output{
		Direction:   direction,
		Strength:    strength,
		Confidence:  confidence,
		Explanation: explanation,
		KeyFactors:  keyFactors,
	}
}

func scoreEvent(input Input) (float64, []scoredFactor) {
	if input.Event == nil && input.News == nil {
		return 0, []scoredFactor{{label: "No linked event was found, so event contribution is neutral", contribution: 0}}
	}

	var score float64
	factors := make([]scoredFactor, 0)

	if input.Event != nil {
		switch input.Event.Type {
		case "key_rate_cut":
			score += 0.35
			factors = append(factors, scoredFactor{label: "Key rate cut supports a positive market reaction", contribution: 0.35})
		case "key_rate_hike":
			score -= 0.35
			factors = append(factors, scoredFactor{label: "Key rate hike creates downside pressure", contribution: -0.35})
		case "key_rate_hold":
			score -= 0.05
			factors = append(factors, scoredFactor{label: "Key rate hold keeps the signal close to neutral", contribution: -0.05})
		case "dividend":
			score += 0.22
			factors = append(factors, scoredFactor{label: "Dividend event adds a positive corporate impulse", contribution: 0.22})
		case "financial_results":
			score += 0.10
			factors = append(factors, scoredFactor{label: "Financial results add a moderate fundamental impulse", contribution: 0.10})
		case "sanctions":
			score -= 0.30
			factors = append(factors, scoredFactor{label: "Sanctions-related event increases downside risk", contribution: -0.30})
		case "commodity_oil":
			if input.Asset.Sector == "oil_gas" || input.Asset.Ticker == "IMOEX" {
				score += 0.14
				factors = append(factors, scoredFactor{label: "Oil-related event supports oil and gas exposure", contribution: 0.14})
			}
		case "commodity_gas":
			if input.Asset.Sector == "oil_gas" {
				score += 0.10
				factors = append(factors, scoredFactor{label: "Gas-related event is relevant for oil and gas exposure", contribution: 0.10})
			}
		case "monetary_policy":
			score -= 0.08
			factors = append(factors, scoredFactor{label: "Monetary policy event adds cautious market tone", contribution: -0.08})
		}
	}

	text := normalizeText(strings.Join([]string{
		input.EventSummary(),
		input.NewsTitle(),
		input.NewsSummary(),
	}, " "))

	positiveHits := countHits(text, "GROWTH", "INCREASE", "RISE", "STRONG", "SUPPORT", "PROFIT", "DIVIDEND", "RECOVERY")
	negativeHits := countHits(text, "DECLINE", "DROP", "FALL", "LOSS", "SANCTION", "PRESSURE", "RISK", "CRISIS", "HIKE", "TIGHTENING")

	switch {
	case positiveHits > negativeHits:
		boost := clamp(0.04*float64(positiveHits-negativeHits), 0, 0.12)
		score += boost
		factors = append(factors, scoredFactor{label: "News wording is tilted to positive market markers", contribution: boost})
	case negativeHits > positiveHits:
		penalty := clamp(0.04*float64(negativeHits-positiveHits), 0, 0.12)
		score -= penalty
		factors = append(factors, scoredFactor{label: "News wording is tilted to negative market markers", contribution: -penalty})
	}

	return clamp(score, -0.45, 0.45), factors
}

func scoreTechnical(input Input) (float64, []scoredFactor) {
	score := 0.0
	factors := make([]scoredFactor, 0)

	if input.Indicators.WeeklyReturn != nil {
		switch {
		case *input.Indicators.WeeklyReturn >= 0.03:
			score += 0.12
			factors = append(factors, scoredFactor{label: "Positive weekly return supports continuation", contribution: 0.12})
		case *input.Indicators.WeeklyReturn <= -0.03:
			score -= 0.12
			factors = append(factors, scoredFactor{label: "Negative weekly return weakens short-term setup", contribution: -0.12})
		}
	}

	if input.Indicators.TrendDirection != nil {
		switch *input.Indicators.TrendDirection {
		case "up":
			score += 0.10
			factors = append(factors, scoredFactor{label: "Trend direction is upward", contribution: 0.10})
		case "down":
			score -= 0.10
			factors = append(factors, scoredFactor{label: "Trend direction is downward", contribution: -0.10})
		}
	}

	if input.Indicators.RSI != nil {
		switch {
		case *input.Indicators.RSI < 30:
			score += 0.06
			factors = append(factors, scoredFactor{label: "Oversold RSI favors a rebound", contribution: 0.06})
		case *input.Indicators.RSI > 70:
			score -= 0.06
			factors = append(factors, scoredFactor{label: "Overbought RSI limits upside", contribution: -0.06})
		}
	}

	if input.Indicators.ChannelPosition != nil && input.Indicators.TrendDirection != nil {
		switch {
		case *input.Indicators.TrendDirection == "up" && *input.Indicators.ChannelPosition > 0.70:
			score += 0.04
			factors = append(factors, scoredFactor{label: "Price is strong within the local range during an uptrend", contribution: 0.04})
		case *input.Indicators.TrendDirection == "down" && *input.Indicators.ChannelPosition < 0.30:
			score -= 0.04
			factors = append(factors, scoredFactor{label: "Price is weak within the local range during a downtrend", contribution: -0.04})
		}
	}

	return clamp(score, -0.30, 0.30), factors
}

func scoreMarket(input Input) (float64, []scoredFactor) {
	switch input.Market.Label {
	case "stable":
		return 0.04, []scoredFactor{{label: "Stable market context modestly supports risk appetite", contribution: 0.04}}
	case "cautious":
		return -0.02, []scoredFactor{{label: "Cautious market context trims upside conviction", contribution: -0.02}}
	case "stressed":
		return -0.06, []scoredFactor{{label: "Stressed market context reduces confidence in upside scenarios", contribution: -0.06}}
	case "crisis":
		return -0.10, []scoredFactor{{label: "Crisis market context materially shifts the balance to defense", contribution: -0.10}}
	default:
		return 0, []scoredFactor{{label: "Market context is neutral because regime data is limited", contribution: 0}}
	}
}

func calculateConfidence(input Input, direction string, eventScore, technicalScore float64) float64 {
	confidence := 0.35

	if input.Event != nil {
		confidence += 0.20
	}
	if input.News != nil {
		confidence += 0.05
	}

	switch input.Indicators.CalculationStatus {
	case "ready":
		confidence += 0.20
	case "partial":
		confidence += 0.10
	}

	if input.Market.Label != "" {
		confidence += 0.10
	}

	if input.Indicators.Volatility != nil && *input.Indicators.Volatility > 0.45 {
		confidence -= 0.10
	}

	if eventScore != 0 && technicalScore != 0 && eventScore*technicalScore < 0 {
		confidence -= 0.15
	}

	if direction == "neutral" {
		confidence -= 0.05
	}

	return clamp(confidence, 0.05, 0.95)
}

func buildExplanation(input Input, direction string, keyFactors []string) string {
	base := "Forecast expects a neutral one-week reaction"
	switch direction {
	case "up":
		base = "Forecast expects an upward one-week reaction"
	case "down":
		base = "Forecast expects a downward one-week reaction"
	}

	parts := []string{base + " for " + input.Asset.Ticker + "."}
	if len(keyFactors) > 0 {
		parts = append(parts, "Key drivers: "+strings.Join(keyFactors, "; ")+".")
	}
	if input.Event == nil {
		parts = append(parts, "No explicit linked event was found, so the signal relies mostly on technical and market context.")
	} else {
		parts = append(parts, "The event context is combined with technical indicators and overall market state rather than interpreted in isolation.")
	}

	return strings.Join(parts, " ")
}

func deriveDirection(score float64) string {
	switch {
	case score >= 0.15:
		return "up"
	case score <= -0.15:
		return "down"
	default:
		return "neutral"
	}
}

func pickTopFactors(factors []scoredFactor, limit int) []string {
	sort.SliceStable(factors, func(i, j int) bool {
		return math.Abs(factors[i].contribution) > math.Abs(factors[j].contribution)
	})

	seen := make(map[string]struct{}, limit)
	result := make([]string, 0, limit)
	for _, factor := range factors {
		label := strings.TrimSpace(factor.label)
		if label == "" {
			continue
		}
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		result = append(result, label)
		if len(result) == limit {
			break
		}
	}

	return result
}

func normalizeText(value string) string {
	value = strings.ToUpper(value)
	replacer := strings.NewReplacer(
		".", " ",
		",", " ",
		":", " ",
		";", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"/", " ",
		"\\", " ",
		"-", " ",
		"_", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	)

	return strings.Join(strings.Fields(replacer.Replace(value)), " ")
}

func countHits(value string, variants ...string) int {
	var hits int
	for _, variant := range variants {
		if strings.Contains(value, variant) {
			hits++
		}
	}
	return hits
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
	return math.Round(value*100) / 100
}

func (i Input) EventSummary() string {
	if i.Event == nil {
		return ""
	}
	return i.Event.Summary
}

func (i Input) NewsTitle() string {
	if i.News == nil {
		return ""
	}
	return i.News.Title
}

func (i Input) NewsSummary() string {
	if i.News == nil {
		return ""
	}
	return i.News.Summary
}

func (p PreparedRequest) MarshalJSON() ([]byte, error) {
	type alias PreparedRequest
	return json.Marshal(alias(p))
}
