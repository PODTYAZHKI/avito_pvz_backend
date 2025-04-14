package repository

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"avito-pvz/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type userRepository struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
	logger  *zerolog.Logger
}

func NewUserRepository(pool *pgxpool.Pool, logger *zerolog.Logger) UserRepository {
	return &userRepository{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger,
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query, args, err := r.builder.
		Insert("users").
		Columns("id", "email", "password_hash", "role").
		Values(user.ID, user.Email, user.Password, user.Role).
		ToSql()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to build query")
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to execute query")
		return err
	}

	r.logger.Info().Msgf("User created with ID: %s", user.ID)
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query, args, err := r.builder.
		Select("id", "email", "password_hash", "role").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to build query")
		return nil, err
	}

	var user models.User
	err = r.pool.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Info().Msgf("User with email %s not found", email)
			return nil, nil
		}
		r.logger.Error().Err(err).Msg("Failed to execute query")
		return nil, err
	}

	r.logger.Info().Msgf("User found with ID: %s", user.ID)
	return &user, nil
}
