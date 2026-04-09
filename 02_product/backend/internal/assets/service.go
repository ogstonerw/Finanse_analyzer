package assets

import (
	"context"
	"errors"

	"diploma-market-ai/02_product/backend/internal/storage"
)

var ErrAssetNotFound = errors.New("asset not found")

type Service struct {
	repository *storage.AssetsRepository
}

type Asset struct {
	ID       string `json:"id"`
	Ticker   string `json:"ticker"`
	Name     string `json:"name"`
	Type     string `json:"asset_type"`
	Sector   string `json:"sector"`
	Currency string `json:"currency"`
	IsActive bool   `json:"is_active"`
}

func NewService(store *storage.Postgres) *Service {
	return &Service{repository: storage.NewAssetsRepository(store)}
}

func (s *Service) List(ctx context.Context) ([]Asset, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Asset, 0, len(items))
	for _, item := range items {
		result = append(result, mapAsset(item))
	}

	return result, nil
}

func (s *Service) GetByTicker(ctx context.Context, ticker string) (Asset, error) {
	item, err := s.repository.GetByTicker(ctx, ticker)
	if err != nil {
		if errors.Is(err, storage.ErrAssetNotFound) {
			return Asset{}, ErrAssetNotFound
		}
		return Asset{}, err
	}

	return mapAsset(item), nil
}

func mapAsset(item storage.AssetRecord) Asset {
	return Asset{
		ID:       item.ID,
		Ticker:   item.Ticker,
		Name:     item.Name,
		Type:     item.AssetType,
		Sector:   item.Sector,
		Currency: item.Currency,
		IsActive: item.IsActive,
	}
}
