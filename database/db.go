package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"avito-pvz/internal/config"
)

func ConnectDB(cfg *config.Config, log *zerolog.Logger) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSslMode,
	)

	
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse database config")
		return nil, err
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime

	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	
	if err := db.Ping(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to ping database")
		return nil, err
	}

	log.Info().Msg("Connected to database")
	return db, nil
}

func CloseDB(db *pgxpool.Pool, log *zerolog.Logger) {
	if db != nil {
		db.Close()
		log.Info().Msg("Database connection closed")
	}
}
