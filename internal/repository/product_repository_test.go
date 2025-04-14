package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"avito-pvz/internal/models"
)

func TestProductRepository_CreateProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		logger := zerolog.Nop()

		repo := &productRepository{
			pool:    mock,
			builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
			logger:  &logger,
		}

		product := &models.Product{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.ProductTypeClothing,
			ReceptionID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO products").
			WithArgs(product.ID, product.DateTime, product.Type, product.ReceptionID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.CreateProduct(context.Background(), product)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		logger := zerolog.Nop()
		repo := &productRepository{
			pool:    mock,
			builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
			logger:  &logger,
		}

		product := &models.Product{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.ProductTypeClothing,
			ReceptionID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO products").
			WithArgs(product.ID, product.DateTime, product.Type, product.ReceptionID).
			WillReturnError(assert.AnError)

		err = repo.CreateProduct(context.Background(), product)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductRepository_GetLastProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		logger := zerolog.Nop()
		repo := &productRepository{
			pool:    mock,
			builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
			logger:  &logger,
		}

		pvzID := uuid.New()
		expected := &models.Product{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        models.ProductTypeClothing,
			ReceptionID: uuid.New(),
		}

		rows := mock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
			AddRow(expected.ID, expected.DateTime, expected.Type, expected.ReceptionID)

		mock.ExpectQuery("SELECT p.id, p.date_time, p.type, p.reception_id FROM products p").
			WithArgs(pvzID.String(), models.ReceptionStatusInProgress).
			WillReturnRows(rows)

		result, err := repo.GetLastProduct(context.Background(), pvzID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		logger := zerolog.Nop()
		repo := &productRepository{
			pool:    mock,
			builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
			logger:  &logger,
		}

		pvzID := uuid.New()

		mock.ExpectQuery("SELECT p.id, p.date_time, p.type, p.reception_id FROM products p").
			WithArgs(pvzID.String(), models.ReceptionStatusInProgress).
			WillReturnError(pgx.ErrNoRows)

		result, err := repo.GetLastProduct(context.Background(), pvzID)
		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		logger := zerolog.Nop()
		repo := &productRepository{
			pool:    mock,
			builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
			logger:  &logger,
		}

		pvzID := uuid.New()

		mock.ExpectQuery("SELECT p.id, p.date_time, p.type, p.reception_id FROM products p").
			WithArgs(pvzID.String(), models.ReceptionStatusInProgress).
			WillReturnError(assert.AnError)

		result, err := repo.GetLastProduct(context.Background(), pvzID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProductRepository_DeleteProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		logger := zerolog.Nop()
		repo := &productRepository{
			pool:    mock,
			builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
			logger:  &logger,
		}

		productID := uuid.New()

		mock.ExpectExec("DELETE FROM products").
			WithArgs(productID.String()).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.DeleteProduct(context.Background(), productID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		logger := zerolog.Nop()
		repo := &productRepository{
			pool:    mock,
			builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
			logger:  &logger,
		}

		productID := uuid.New()

		mock.ExpectExec("DELETE FROM products").
			WithArgs(productID.String()).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = repo.DeleteProduct(context.Background(), productID)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		logger := zerolog.Nop()
		repo := &productRepository{
			pool:    mock,
			builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
			logger:  &logger,
		}

		productID := uuid.New()

		mock.ExpectExec("DELETE FROM products").
			WithArgs(productID.String()).
			WillReturnError(assert.AnError)

		err = repo.DeleteProduct(context.Background(), productID)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
