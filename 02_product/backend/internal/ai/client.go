package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	ModeFallback = "fallback"
	ModePrepare  = "prepare"
	ModeOpenAI   = "openai"

	defaultOpenAIEndpoint = "https://api.openai.com/v1/responses"
	defaultOpenAIModel    = "gpt-5.2"
	defaultHTTPTimeoutSec = 30
)

type Config struct {
	Mode            string
	Provider        string
	Model           string
	Endpoint        string
	APIKey          string
	ReasoningEffort string
	TextVerbosity   string
	TimeoutSeconds  int
}

type Client struct {
	config     Config
	httpClient *http.Client
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

type openAIResponseRequest struct {
	Model        string           `json:"model"`
	Instructions string           `json:"instructions"`
	Input        string           `json:"input"`
	Text         openAITextConfig `json:"text"`
	Reasoning    *openAIReasoning `json:"reasoning,omitempty"`
}

type openAITextConfig struct {
	Format    openAITextFormat `json:"format"`
	Verbosity string           `json:"verbosity,omitempty"`
}

type openAITextFormat struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Strict      bool           `json:"strict"`
	Schema      map[string]any `json:"schema"`
}

type openAIReasoning struct {
	Effort string `json:"effort,omitempty"`
}

type openAIResponsePayload struct {
	ID         string              `json:"id"`
	OutputText string              `json:"output_text"`
	Output     []openAIOutputItem  `json:"output"`
	Error      *openAIErrorPayload `json:"error,omitempty"`
}

type openAIOutputItem struct {
	Type    string              `json:"type"`
	Content []openAIContentItem `json:"content"`
}

type openAIContentItem struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	Refusal string `json:"refusal"`
}

type openAIErrorPayload struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    any    `json:"code"`
}

type structuredForecast struct {
	Direction   string   `json:"direction"`
	Strength    float64  `json:"strength"`
	Confidence  float64  `json:"confidence"`
	Explanation string   `json:"explanation"`
	KeyFactors  []string `json:"key_factors"`
}

func NewClient(cfg Config) *Client {
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	switch mode {
	case ModeFallback, ModePrepare, ModeOpenAI:
	default:
		mode = ModeFallback
	}

	provider := strings.TrimSpace(cfg.Provider)
	if provider == "" {
		provider = "openai"
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		if mode == ModeFallback {
			model = "fallback-rule-engine"
		} else {
			model = defaultOpenAIModel
		}
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" && mode != ModeFallback {
		endpoint = defaultOpenAIEndpoint
	}

	timeoutSeconds := cfg.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultHTTPTimeoutSec
	}

	return &Client{
		config: Config{
			Mode:            mode,
			Provider:        provider,
			Model:           model,
			Endpoint:        endpoint,
			APIKey:          strings.TrimSpace(cfg.APIKey),
			ReasoningEffort: strings.ToLower(strings.TrimSpace(cfg.ReasoningEffort)),
			TextVerbosity:   strings.ToLower(strings.TrimSpace(cfg.TextVerbosity)),
			TimeoutSeconds:  timeoutSeconds,
		},
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

func (c *Client) Generate(ctx context.Context, input Input) (Output, error) {
	if c.config.Mode == ModeOpenAI {
		return c.generateOpenAI(ctx, input)
	}

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
	payload := c.buildOpenAIRequestPayload(input)

	return PreparedRequest{
		Provider: c.config.Provider,
		Endpoint: c.config.Endpoint,
		Model:    c.config.Model,
		Payload:  payload,
	}
}

func (c *Client) generateOpenAI(ctx context.Context, input Input) (Output, error) {
	if strings.ToLower(c.config.Provider) != "openai" {
		return Output{}, fmt.Errorf("unsupported ai provider for openai mode: %s", c.config.Provider)
	}
	if c.config.APIKey == "" {
		return Output{}, errors.New("AI_API_KEY is required when AI_MODE=openai")
	}
	if c.config.Endpoint == "" {
		return Output{}, errors.New("AI_API_ENDPOINT is required when AI_MODE=openai")
	}

	prepared := c.BuildPreparedRequest(input)
	body, err := json.Marshal(prepared.Payload)
	if err != nil {
		return Output{}, fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return Output{}, fmt.Errorf("create openai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Output{}, fmt.Errorf("send openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Output{}, fmt.Errorf("read openai response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Output{}, fmt.Errorf("openai response status %s: %s", resp.Status, openAIErrorMessage(respBody))
	}

	var payload openAIResponsePayload
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return Output{}, fmt.Errorf("decode openai response: %w", err)
	}
	if payload.Error != nil && strings.TrimSpace(payload.Error.Message) != "" {
		return Output{}, fmt.Errorf("openai response error: %s", payload.Error.Message)
	}

	outputText, err := extractOpenAIOutputText(payload)
	if err != nil {
		return Output{}, err
	}

	var parsed structuredForecast
	if err := json.Unmarshal([]byte(outputText), &parsed); err != nil {
		return Output{}, fmt.Errorf("decode structured forecast output: %w", err)
	}

	output, err := normalizeStructuredForecast(parsed)
	if err != nil {
		return Output{}, err
	}
	output.Mode = ModeOpenAI
	output.Model = c.config.Model
	output.PreparedRequest = &prepared

	return output, nil
}

func (c *Client) buildOpenAIRequestPayload(input Input) map[string]any {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		inputJSON = []byte("{}")
	}

	request := openAIResponseRequest{
		Model:        c.config.Model,
		Instructions: buildOpenAIInstructions(),
		Input:        "Structured market forecast context as JSON:\n" + string(inputJSON),
		Text: openAITextConfig{
			Format: openAITextFormat{
				Type:        "json_schema",
				Name:        "market_forecast_signal",
				Description: "One-week market reaction forecast generated from structured market, news, event and technical context.",
				Strict:      true,
				Schema:      forecastOutputSchema(),
			},
			Verbosity: c.config.TextVerbosity,
		},
	}

	if c.config.ReasoningEffort != "" {
		request.Reasoning = &openAIReasoning{Effort: c.config.ReasoningEffort}
	}

	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return map[string]any{}
	}

	payload := make(map[string]any)
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return map[string]any{}
	}

	return payload
}

func buildOpenAIInstructions() string {
	return strings.Join([]string{
		"You are the AI forecasting module of a Russian market analytics MVP.",
		"Use only the structured input provided by the backend. Do not invent market data, sources, prices, events, or metrics.",
		"Generate a one-week reaction forecast for the requested asset.",
		"Interpret news and event text together with technical indicators and market regime; do not base the signal only on raw news sentiment.",
		"Return direction as exactly one of: up, neutral, down.",
		"Return strength and confidence as normalized numbers in the 0..1 range.",
		"Write explanation and key_factors in Russian, briefly and without investment advice.",
		"If input data is incomplete or contradictory, lower confidence and explain the limitation.",
	}, " ")
}

func forecastOutputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"direction": map[string]any{
				"type":        "string",
				"description": "Expected one-week market reaction direction.",
				"enum":        []string{"up", "neutral", "down"},
			},
			"strength": map[string]any{
				"type":        "number",
				"description": "Normalized signal strength from 0 to 1.",
			},
			"confidence": map[string]any{
				"type":        "number",
				"description": "Normalized confidence score from 0 to 1.",
			},
			"explanation": map[string]any{
				"type":        "string",
				"description": "Short Russian explanation of the forecast and its limits.",
			},
			"key_factors": map[string]any{
				"type":        "array",
				"description": "Main Russian-language factors that influenced the signal.",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{
			"direction",
			"strength",
			"confidence",
			"explanation",
			"key_factors",
		},
	}
}

func openAIErrorMessage(body []byte) string {
	var payload openAIResponsePayload
	if err := json.Unmarshal(body, &payload); err == nil && payload.Error != nil && payload.Error.Message != "" {
		return payload.Error.Message
	}

	message := strings.TrimSpace(string(body))
	if len(message) > 500 {
		message = message[:500]
	}
	if message == "" {
		message = "empty response body"
	}

	return message
}

func extractOpenAIOutputText(payload openAIResponsePayload) (string, error) {
	if text := strings.TrimSpace(payload.OutputText); text != "" {
		return text, nil
	}

	var refusal string
	for _, item := range payload.Output {
		for _, content := range item.Content {
			if text := strings.TrimSpace(content.Text); text != "" {
				return text, nil
			}
			if value := strings.TrimSpace(content.Refusal); value != "" {
				refusal = value
			}
		}
	}

	if refusal != "" {
		return "", fmt.Errorf("openai refused forecast request: %s", refusal)
	}

	return "", errors.New("openai response did not contain output text")
}

func normalizeStructuredForecast(parsed structuredForecast) (Output, error) {
	direction := strings.ToLower(strings.TrimSpace(parsed.Direction))
	switch direction {
	case "up", "neutral", "down":
	default:
		return Output{}, fmt.Errorf("invalid forecast direction from openai: %s", parsed.Direction)
	}

	explanation := strings.TrimSpace(parsed.Explanation)
	if explanation == "" {
		return Output{}, errors.New("openai forecast explanation is empty")
	}

	keyFactors := normalizeKeyFactors(parsed.KeyFactors, 6)

	return Output{
		Direction:   direction,
		Strength:    round2(clamp(parsed.Strength, 0, 1)),
		Confidence:  round2(clamp(parsed.Confidence, 0, 1)),
		Explanation: explanation,
		KeyFactors:  keyFactors,
	}, nil
}

func normalizeKeyFactors(values []string, limit int) []string {
	if limit <= 0 {
		limit = len(values)
	}

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
		if len(result) == limit {
			break
		}
	}

	return result
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
