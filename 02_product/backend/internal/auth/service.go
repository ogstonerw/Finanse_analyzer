package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"diploma-market-ai/02_product/backend/internal/users"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidInput       = errors.New("email and password are required")
)

type Service struct {
	usersRepo  *users.Repository
	sessionTTL time.Duration
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResult struct {
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewService(usersRepo *users.Repository, sessionTTL time.Duration) *Service {
	if sessionTTL <= 0 {
		sessionTTL = 24 * time.Hour
	}

	return &Service{
		usersRepo:  usersRepo,
		sessionTTL: sessionTTL,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterRequest, meta users.SessionMeta) (AuthResult, error) {
	email := normalizeEmail(input.Email)
	if email == "" || input.Password == "" {
		return AuthResult{}, ErrInvalidInput
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return AuthResult{}, err
	}

	user, err := s.usersRepo.CreateUser(ctx, users.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if errors.Is(err, users.ErrUserAlreadyExists) {
			return AuthResult{}, ErrEmailAlreadyExists
		}
		return AuthResult{}, err
	}

	session, err := s.createSession(ctx, user.ID, meta)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		UserID:       user.ID,
		Email:        user.Email,
		SessionToken: session.SessionToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginRequest, meta users.SessionMeta) (AuthResult, error) {
	email := normalizeEmail(input.Email)
	if email == "" || input.Password == "" {
		return AuthResult{}, ErrInvalidInput
	}

	user, err := s.usersRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	if err := comparePassword(user.PasswordHash, input.Password); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	if err := s.usersRepo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return AuthResult{}, err
	}

	session, err := s.createSession(ctx, user.ID, meta)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		UserID:       user.ID,
		Email:        user.Email,
		SessionToken: session.SessionToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

func (s *Service) createSession(ctx context.Context, userID string, meta users.SessionMeta) (users.UserSession, error) {
	token, err := generateToken()
	if err != nil {
		return users.UserSession{}, err
	}

	expiresAt := time.Now().UTC().Add(s.sessionTTL)

	return s.usersRepo.CreateSession(ctx, users.CreateSessionParams{
		UserID:       userID,
		SessionToken: token,
		IP:           meta.IP,
		UserAgent:    meta.UserAgent,
		ExpiresAt:    expiresAt,
	})
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
