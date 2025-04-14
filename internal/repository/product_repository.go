package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"

	"avito-pvz/internal/models"
)

type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetLastProduct(ctx context.Context, pvzID uuid.UUID) (*models.Product, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
}

type productRepository struct {
	pool    PgxPool
	builder squirrel.StatementBuilderType
	logger  *zerolog.Logger
}

func NewProductRepository(pool PgxPool, logger *zerolog.Logger) ProductRepository {
	return &productRepository{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger,
	}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	query, args, err := r.builder.
		Insert("products").
		Columns("id", "date_time", "type", "reception_id").
		Values(
			product.ID,
			product.DateTime,
			product.Type,
			product.ReceptionID,
		).
		ToSql()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to build SQL query")
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("productID", product.ID.String()).
			Msg("Failed to create product")
		return err
	}

	return nil
}

func (r *productRepository) GetLastProduct(ctx context.Context, pvzID uuid.UUID) (*models.Product, error) {
	query, args, err := r.builder.
		Select("p.id", "p.date_time", "p.type", "p.reception_id").
		From("products p").
		Join("receptions r ON p.reception_id = r.id").
		Where(squirrel.Eq{
			"r.pvz_id": pvzID,
			"r.status": models.ReceptionStatusInProgress,
		}).
		OrderBy("p.date_time DESC").
		Limit(1).
		ToSql()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to build SQL query")
		return nil, err
	}

	var product models.Product
	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&product.ID,
		&product.DateTime,
		&product.Type,
		&product.ReceptionID,
	)

	if err == pgx.ErrNoRows {
		r.logger.Debug().Str("pvzID", pvzID.String()).Msg("No products found")
		return nil, nil
	}

	if err != nil {
		r.logger.Error().Err(err).Str("pvzID", pvzID.String()).Msg("Failed to get last product")
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	query, args, err := r.builder.
		Delete("products").
		Where(squirrel.Eq{"id": productID}).
		ToSql()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to build SQL query")
		return err
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("productID", productID.String()).Msg("Failed to delete product")
		return err
	}

	if result.RowsAffected() == 0 {
		r.logger.Warn().Str("productID", productID.String()).Msg("Product not found")
		return pgx.ErrNoRows
	}

	return nil
}
