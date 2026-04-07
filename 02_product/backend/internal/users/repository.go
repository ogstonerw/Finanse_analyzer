package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"diploma-market-ai/02_product/backend/internal/storage"
	"github.com/lib/pq"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  sql.NullTime
}

type UserSession struct {
	ID           string
	UserID       string
	SessionToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
}

type CreateSessionParams struct {
	UserID       string
	SessionToken string
	IP           string
	UserAgent    string
	ExpiresAt    time.Time
}

type SessionMeta struct {
	IP        string
	UserAgent string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(store *storage.Postgres) *Repository {
	return &Repository{db: store.DB()}
}

func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams) (User, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, role, is_active, created_at, updated_at, last_login_at
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, params.Email, params.PasswordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrUserAlreadyExists
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT id, email, password_hash, role, is_active, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, userID string, loggedAt time.Time) error {
	const query = `
		UPDATE users
		SET last_login_at = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, loggedAt)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	return nil
}

func (r *Repository) CreateSession(ctx context.Context, params CreateSessionParams) (UserSession, error) {
	const query = `
		INSERT INTO user_sessions (user_id, session_token, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, session_token, expires_at, created_at
	`

	var session UserSession
	err := r.db.QueryRowContext(
		ctx,
		query,
		params.UserID,
		params.SessionToken,
		params.IP,
		params.UserAgent,
		params.ExpiresAt,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.SessionToken,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		return UserSession{}, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return string(pgErr.Code) == "23505"
	}
	return false
}
