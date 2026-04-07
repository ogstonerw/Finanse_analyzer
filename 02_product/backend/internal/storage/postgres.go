package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type Postgres struct {
	config Config
	dsn    string
	db     *sql.DB
}

func NewPostgres(cfg Config) (*Postgres, error) {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Name == "" {
		return nil, errors.New("postgres config is incomplete")
	}

	dsn := buildDSN(cfg)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Postgres{
		config: cfg,
		dsn:    dsn,
		db:     db,
	}, nil
}

func (p *Postgres) DB() *sql.DB {
	return p.db
}

func (p *Postgres) Close() {
	if p.db != nil {
		_ = p.db.Close()
	}
}

func (p *Postgres) DSN() string {
	return p.dsn
}

func (p *Postgres) DriverName() string {
	return "postgres"
}

func buildDSN(cfg Config) string {
	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?%s",
		url.QueryEscape(cfg.User),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		cfg.Name,
		query.Encode(),
	)
}
