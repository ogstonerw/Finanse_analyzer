package storage

import (
	"errors"
	"fmt"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Postgres struct {
	config Config
	dsn    string
}

func NewPostgres(cfg Config) (*Postgres, error) {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Name == "" {
		return nil, errors.New("postgres config is incomplete")
	}

	return &Postgres{
		config: cfg,
		dsn:    buildDSN(cfg),
	}, nil
}

func (p *Postgres) DSN() string {
	return p.dsn
}

func (p *Postgres) DriverName() string {
	return "postgres-placeholder"
}

func buildDSN(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)
}
