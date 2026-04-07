package auth

import (
	"context"
	"errors"

	"diploma-market-ai/02_product/backend/internal/storage"
)

var ErrNotImplemented = errors.New("auth business logic is not implemented yet")

type Service struct {
	store *storage.Postgres
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewService(store *storage.Postgres) *Service {
	return &Service{store: store}
}

func (s *Service) Register(ctx context.Context, input RegisterRequest) error {
	_ = ctx
	_ = input
	return ErrNotImplemented
}

func (s *Service) Login(ctx context.Context, input LoginRequest) error {
	_ = ctx
	_ = input
	return ErrNotImplemented
}
