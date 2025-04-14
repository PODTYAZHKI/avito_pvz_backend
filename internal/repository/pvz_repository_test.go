package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"avito-pvz/internal/models"
)

func TestPvzRepository_CreatePvz(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zerolog.Nop()
	repo := &pvzRepository{
		pool:    mock,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  &logger,
	}

	t.Run("success", func(t *testing.T) {
		pvz := &models.Pvz{
			ID:               uuid.New(),
			RegistrationDate: time.Now(),
			City:             models.Moscow,
		}

		mock.ExpectExec("INSERT INTO pvz").
			WithArgs(pvz.ID, pvz.RegistrationDate, pvz.City).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.CreatePvz(context.Background(), pvz)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		pvz := &models.Pvz{
			ID:               uuid.New(),
			RegistrationDate: time.Now(),
			City:             models.Moscow,
		}

		mock.ExpectExec("INSERT INTO pvz").
			WithArgs(pvz.ID, pvz.RegistrationDate, pvz.City).
			WillReturnError(fmt.Errorf("error"))

		err = repo.CreatePvz(context.Background(), pvz)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPvzRepository_GetPVZsWithReceptions(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zerolog.Nop()
	repo := &pvzRepository{
		pool:    mock,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  &logger,
	}

	testPVZ := models.Pvz{
		ID:               uuid.New(),
		City:             models.Moscow,
		RegistrationDate: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		filter := models.PvzFilter{
			Page:  1,
			Limit: 10,
		}

		mock.ExpectQuery("SELECT id, city, registration_date FROM pvz LIMIT 10 OFFSET 0").
			WillReturnRows(mock.NewRows([]string{"id", "city", "registration_date"}).
				AddRow(testPVZ.ID, testPVZ.City, testPVZ.RegistrationDate))

		mock.ExpectQuery("SELECT id, date_time, status, pvz_id FROM receptions ORDER BY pvz_id, date_time DESC").
			WillReturnRows(mock.NewRows([]string{"id", "date_time", "status", "pvz_id"}).
				AddRow(uuid.New(), time.Now(), models.ReceptionStatusInProgress, testPVZ.ID))

		mock.ExpectQuery("SELECT id, date_time, type, reception_id FROM products ORDER BY reception_id, date_time DESC").
			WillReturnRows(mock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
				AddRow(uuid.New(), time.Now(), models.ProductTypeElectronics, uuid.New()))

		results, err := repo.GetPVZsWithReceptions(context.Background(), filter)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, testPVZ.ID, results[0].PVZ.ID)
		assert.Equal(t, testPVZ.City, results[0].PVZ.City)
	})

	t.Run("error on get paginated pvzs", func(t *testing.T) {
		filter := models.PvzFilter{
			Page:  1,
			Limit: 10,
		}

		mock.ExpectQuery("SELECT id, city, registration_date FROM pvz LIMIT 10 OFFSET 0").
			WillReturnError(fmt.Errorf("error getting PVZs"))

		results, err := repo.GetPVZsWithReceptions(context.Background(), filter)
		assert.Error(t, err)
		assert.Empty(t, results)
	})

	t.Run("error on get receptions", func(t *testing.T) {
		filter := models.PvzFilter{
			Page:  1,
			Limit: 10,
		}

		mock.ExpectQuery("SELECT id, city, registration_date FROM pvz LIMIT 10 OFFSET 0").
			WillReturnRows(mock.NewRows([]string{"id", "city", "registration_date"}).
				AddRow(testPVZ.ID, testPVZ.City, testPVZ.RegistrationDate))

		mock.ExpectQuery("SELECT id, date_time, status, pvz_id FROM receptions ORDER BY pvz_id, date_time DESC").
			WillReturnError(fmt.Errorf("error getting receptions"))

		results, err := repo.GetPVZsWithReceptions(context.Background(), filter)
		assert.Error(t, err)
		assert.Empty(t, results)
	})

	t.Run("error on get products", func(t *testing.T) {
		filter := models.PvzFilter{
			Page:  1,
			Limit: 10,
		}

		mock.ExpectQuery("SELECT id, city, registration_date FROM pvz LIMIT 10 OFFSET 0").
			WillReturnRows(mock.NewRows([]string{"id", "city", "registration_date"}).
				AddRow(testPVZ.ID, testPVZ.City, testPVZ.RegistrationDate))

		mock.ExpectQuery("SELECT id, date_time, status, pvz_id FROM receptions ORDER BY pvz_id, date_time DESC").
			WillReturnRows(mock.NewRows([]string{"id", "date_time", "status", "pvz_id"}).
				AddRow(uuid.New(), time.Now(), models.ReceptionStatusInProgress, testPVZ.ID))

		mock.ExpectQuery("SELECT id, date_time, type, reception_id FROM products ORDER BY reception_id, date_time DESC").
			WillReturnError(fmt.Errorf("error getting products"))

		results, err := repo.GetPVZsWithReceptions(context.Background(), filter)
		assert.Error(t, err)
		assert.Empty(t, results)
	})
}
