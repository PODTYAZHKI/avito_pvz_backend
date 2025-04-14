package repository

import (
	"context"
	"fmt"
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

func TestReceptionRepository_CreateReception(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zerolog.Nop()
	repo := &receptionRepository{
		pool:    mock,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  &logger,
	}

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    uuid.New(),
		Status:   models.ReceptionStatusInProgress,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO receptions").
			WithArgs(reception.ID, reception.DateTime, reception.PVZID, reception.Status).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.CreateReception(context.Background(), reception)
		assert.NoError(t, err)
	})

	t.Run("error on exec", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO receptions").
			WithArgs(reception.ID, reception.DateTime, reception.PVZID, reception.Status).
			WillReturnError(fmt.Errorf("database error"))

		err := repo.CreateReception(context.Background(), reception)
		assert.Error(t, err)
	})
}

func TestReceptionRepository_GetActiveReception(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zerolog.Nop()
	repo := &receptionRepository{
		pool:    mock,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  &logger,
	}

	pvzID := uuid.New()
	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	t.Run("success", func(t *testing.T) {

		mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = \\$1 AND status = \\$2").
			WithArgs(pvzID.String(), models.ReceptionStatusInProgress).
			WillReturnRows(mock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
				AddRow(reception.ID, reception.DateTime, reception.PVZID, reception.Status))

		result, err := repo.GetActiveReception(context.Background(), pvzID)
		fmt.Println()
		fmt.Println("result", result)
		fmt.Println()
		fmt.Println()
		fmt.Println("reception", reception)
		fmt.Println()
		assert.NoError(t, err)
		assert.Equal(t, reception, result)
	})

	t.Run("no rows", func(t *testing.T) {

		mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = \\$1 AND status = \\$2").
			WithArgs(pvzID.String(), models.ReceptionStatusInProgress).
			WillReturnError(pgx.ErrNoRows)

		result, err := repo.GetActiveReception(context.Background(), pvzID)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("error on query", func(t *testing.T) {

		mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = \\$1 AND status = \\$2").
			WithArgs(pvzID.String(), models.ReceptionStatusInProgress).
			WillReturnError(fmt.Errorf("database error"))

		result, err := repo.GetActiveReception(context.Background(), pvzID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestReceptionRepository_CloseReception(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zerolog.Nop()
	repo := &receptionRepository{
		pool:    mock,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  &logger,
	}

	receptionID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("UPDATE receptions").
			WithArgs(models.ReceptionStatusClosed, receptionID.String()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.CloseReception(context.Background(), receptionID)
		assert.NoError(t, err)
	})

	t.Run("error on exec", func(t *testing.T) {
		mock.ExpectExec("UPDATE receptions").
			WithArgs(models.ReceptionStatusClosed, receptionID).
			WillReturnError(fmt.Errorf("database error"))

		err := repo.CloseReception(context.Background(), receptionID)
		assert.Error(t, err)
	})
}
