package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateOpenAI(t *testing.T) {
	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")

		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if request["model"] != "test-model" {
			t.Fatalf("unexpected model: %v", request["model"])
		}
		if _, ok := request["text"].(map[string]any); !ok {
			t.Fatalf("text format config is missing")
		}

		output, err := json.Marshal(structuredForecast{
			Direction:   "up",
			Strength:    0.72,
			Confidence:  0.61,
			Explanation: "Context supports an upward reaction.",
			KeyFactors:  []string{"Event", "Trend"},
		})
		if err != nil {
			t.Fatalf("marshal output: %v", err)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "resp_test",
			"output": []map[string]any{
				{
					"type": "message",
					"content": []map[string]any{
						{
							"type": "output_text",
							"text": string(output),
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(Config{
		Mode:           ModeOpenAI,
		Provider:       "openai",
		Model:          "test-model",
		Endpoint:       server.URL,
		APIKey:         "test-key",
		TimeoutSeconds: 5,
	})

	output, err := client.Generate(context.Background(), sampleInput())
	if err != nil {
		t.Fatalf("generate openai: %v", err)
	}

	if capturedAuth != "Bearer test-key" {
		t.Fatalf("unexpected authorization header: %s", capturedAuth)
	}
	if output.Direction != "up" {
		t.Fatalf("unexpected direction: %s", output.Direction)
	}
	if output.Mode != ModeOpenAI {
		t.Fatalf("unexpected mode: %s", output.Mode)
	}
	if output.Model != "test-model" {
		t.Fatalf("unexpected model: %s", output.Model)
	}
	if output.PreparedRequest == nil {
		t.Fatalf("prepared request should be stored")
	}
}

func TestGenerateOpenAIRequiresAPIKey(t *testing.T) {
	client := NewClient(Config{
		Mode:     ModeOpenAI,
		Provider: "openai",
		Model:    "test-model",
		Endpoint: "https://example.test/v1/responses",
	})

	_, err := client.Generate(context.Background(), sampleInput())
	if err == nil {
		t.Fatalf("expected missing api key error")
	}
}

func TestPreparedRequestDoesNotContainAPIKey(t *testing.T) {
	client := NewClient(Config{
		Mode:     ModePrepare,
		Provider: "openai",
		Model:    "test-model",
		Endpoint: "https://example.test/v1/responses",
		APIKey:   "secret-key",
	})

	prepared := client.BuildPreparedRequest(sampleInput())
	raw, err := json.Marshal(prepared)
	if err != nil {
		t.Fatalf("marshal prepared request: %v", err)
	}

	if strings.Contains(string(raw), "secret-key") {
		t.Fatalf("prepared request should not contain api key")
	}
	if !strings.Contains(string(raw), "json_schema") {
		t.Fatalf("prepared request should contain structured output schema")
	}
}

func sampleInput() Input {
	trend := "up"
	weeklyReturn := 0.04
	rsi := 55.0

	return Input{
		Horizon: "1w",
		Asset: AssetContext{
			ID:        "asset-1",
			Ticker:    "SBER",
			Name:      "Sber",
			AssetType: "equity",
			Sector:    "banking",
			Currency:  "RUB",
		},
		Event: &EventContext{
			ID:      "event-1",
			Type:    "financial_results",
			Summary: "The company published financial results.",
		},
		Indicators: IndicatorContext{
			Timeframe:         "1d",
			WeeklyReturn:      &weeklyReturn,
			RSI:               &rsi,
			TrendDirection:    &trend,
			CalculationStatus: "ready",
		},
		Market: MarketContext{
			Label:       "stable",
			Score:       0.20,
			Explanation: "The market is in a stable regime.",
		},
	}
}
