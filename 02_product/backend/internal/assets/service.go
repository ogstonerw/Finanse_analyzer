package assets

import (
	"context"

	"diploma-market-ai/02_product/backend/internal/storage"
)

type Service struct {
	store *storage.Postgres
}

type Asset struct {
	ID       string `json:"id"`
	Ticker   string `json:"ticker"`
	Name     string `json:"name"`
	Type     string `json:"asset_type"`
	IsActive bool   `json:"is_active"`
}

func NewService(store *storage.Postgres) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context) ([]Asset, error) {
	_ = ctx

	return []Asset{
		{Ticker: "IMOEX", Name: "Индекс Московской биржи", Type: "index", IsActive: true},
		{Ticker: "SBER", Name: "Сбер", Type: "equity", IsActive: true},
		{Ticker: "LKOH", Name: "Лукойл", Type: "equity", IsActive: true},
		{Ticker: "GAZP", Name: "Газпром", Type: "equity", IsActive: true},
		{Ticker: "YDEX", Name: "Яндекс", Type: "equity", IsActive: true},
	}, nil
}
