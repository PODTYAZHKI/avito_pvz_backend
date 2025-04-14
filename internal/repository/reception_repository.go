package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"avito-pvz/internal/models"
)

type ReceptionRepository interface {
	CreateReception(ctx context.Context, reception *models.Reception) error
	GetActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, receptionID uuid.UUID) error
}

type receptionRepository struct {
	pool    PgxPool
	builder squirrel.StatementBuilderType
	logger  *zerolog.Logger
}

func NewReceptionRepository(pool PgxPool, logger *zerolog.Logger) ReceptionRepository {
	return &receptionRepository{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger,
	}
}

func (r *receptionRepository) CreateReception(ctx context.Context, reception *models.Reception) error {
	query, args, err := r.builder.Insert("receptions").Columns("id", "date_time", "pvz_id", "status").Values(reception.ID, reception.DateTime, reception.PVZID, reception.Status).ToSql()

	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to build SQL query")
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create reception")
		return err
	}

	return nil
}

func (r *receptionRepository) GetActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	query, args, err := r.builder.
		Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{
			"pvz_id": pvzID,
			"status": models.ReceptionStatusInProgress,
		}).
		Limit(1).
		ToSql()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to build SQL query")
		return nil, err
	}

	var reception models.Reception
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PVZID,
		&reception.Status,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		r.logger.Error().
			Err(err).
			Str("pvzID", pvzID.String()).
			Msg("Failed to get active reception")
		return nil, err
	}

	return &reception, nil
}

func (r *receptionRepository) CloseReception(ctx context.Context, receptionID uuid.UUID) error {
	query, args, err := r.builder.
		Update("receptions").
		Set("status", models.ReceptionStatusClosed).
		Where(squirrel.Eq{"id": receptionID}).
		ToSql()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to build SQL query")
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("receptionID", receptionID.String()).
			Msg("Failed to close reception")
		return err
	}

	return nil
}
