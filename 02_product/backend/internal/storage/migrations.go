package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultMigrationsDir = "migrations"

// ApplyMigrations applies SQL files from the migrations directory once.
// It is intentionally small and transparent for the MVP: every applied file is
// registered in schema_migrations by filename, so local demo databases can be
// prepared by simply starting the backend.
func (p *Postgres) ApplyMigrations(ctx context.Context, migrationsDir string) error {
	if strings.TrimSpace(migrationsDir) == "" {
		migrationsDir = defaultMigrationsDir
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory %s: %w", migrationsDir, err)
	}

	migrationFiles := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		migrationFiles = append(migrationFiles, file.Name())
	}
	sort.Strings(migrationFiles)

	if _, err := p.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	for _, fileName := range migrationFiles {
		applied, err := p.isMigrationApplied(ctx, fileName)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, fileName))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", fileName, err)
		}

		if err := p.applyMigration(ctx, fileName, string(content)); err != nil {
			return err
		}
	}

	return nil
}

func (p *Postgres) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	var exists bool
	err := p.db.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`,
		version,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}

	return exists, nil
}

func (p *Postgres) applyMigration(ctx context.Context, version string, sqlText string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}
	defer rollbackOnError(tx, &err)

	if _, err = tx.ExecContext(ctx, sqlText); err != nil {
		return fmt.Errorf("execute migration %s: %w", version, err)
	}

	if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return fmt.Errorf("register migration %s: %w", version, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}

	return nil
}

func rollbackOnError(tx *sql.Tx, err *error) {
	if err != nil && *err != nil {
		_ = tx.Rollback()
	}
}
