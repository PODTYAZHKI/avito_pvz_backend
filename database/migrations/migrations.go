package migrations

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func RunMigration(ctx context.Context, pool *pgxpool.Pool, log *zerolog.Logger) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current file path")
	}

	migrationFilePath := filepath.Join(filepath.Dir(filename), "000001_init_schema.up.sql")

	sqlBytes, err := os.ReadFile(migrationFilePath)
	if err != nil {
		log.Error().Err(err).Str("path", migrationFilePath).Msg("Failed to read migration file")
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	sqlContent := string(sqlBytes)
	queries := strings.Split(sqlContent, ";")

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Error().Err(err).Msg("Transaction rollback failed")
		}
	}()

	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		query += ";"

		if _, err := tx.Exec(ctx, query); err != nil {
			log.Error().Err(err).Str("query", query).Msg("Failed to execute migration query")
			return fmt.Errorf("failed to execute query '%s': %w", query, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().Msg("Database schema initialized successfully")
	return nil
}
