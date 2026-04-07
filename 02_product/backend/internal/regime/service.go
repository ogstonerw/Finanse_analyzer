package regime

import (
	"context"
	"time"

	"diploma-market-ai/02_product/backend/internal/storage"
)

type Service struct {
	store *storage.Postgres
}

type MarketRegime struct {
	Label        string    `json:"label"`
	Score        float64   `json:"score"`
	CalculatedAt time.Time `json:"calculated_at"`
	Explanation  string    `json:"explanation"`
}

func NewService(store *storage.Postgres) *Service {
	return &Service{store: store}
}

func (s *Service) Current(ctx context.Context) (MarketRegime, error) {
	_ = ctx

	return MarketRegime{
		Label:        "stable",
		Score:        0,
		CalculatedAt: time.Now().UTC(),
		Explanation:  "regime endpoint is scaffolded; hybrid crisisometer logic will be connected later",
	}, nil
}
