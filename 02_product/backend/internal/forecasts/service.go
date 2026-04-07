package forecasts

import (
	"context"
	"time"

	"diploma-market-ai/02_product/backend/internal/storage"
)

type Service struct {
	store *storage.Postgres
}

type Forecast struct {
	Asset       string    `json:"asset"`
	Horizon     string    `json:"horizon"`
	Direction   string    `json:"direction"`
	Strength    string    `json:"strength"`
	Confidence  string    `json:"confidence"`
	Explanation string    `json:"explanation"`
	GeneratedAt time.Time `json:"generated_at"`
	Source      string    `json:"source"`
}

func NewService(store *storage.Postgres) *Service {
	return &Service{store: store}
}

func (s *Service) Latest(ctx context.Context) (Forecast, error) {
	_ = ctx

	return Forecast{
		Asset:       "IMOEX",
		Horizon:     "1w",
		Direction:   "neutral",
		Strength:    "placeholder",
		Confidence:  "placeholder",
		Explanation: "latest forecast endpoint is scaffolded; business logic will be added later",
		GeneratedAt: time.Now().UTC(),
		Source:      s.store.DriverName(),
	}, nil
}
