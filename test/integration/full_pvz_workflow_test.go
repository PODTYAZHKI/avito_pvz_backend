package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"avito-pvz/internal/models"
	"avito-pvz/internal/repository"
	"avito-pvz/internal/service"
)

func TestFullPVZWorkflow(t *testing.T) {

	ctx := context.Background()
	pool := setupTestDB(t)
	defer pool.Close()

	logger := zerolog.New(nil)

	pvzRepo := repository.NewPvzRepository(pool, &logger)
	receptionRepo := repository.NewReceptionRepository(pool, &logger)
	productRepo := repository.NewProductRepository(pool, &logger)

	pvzService := service.NewPvzService(pvzRepo, &logger)
	receptionService := service.NewReceptionService(receptionRepo, &logger)
	productService := service.NewProductService(receptionRepo, productRepo, &logger)

	pvz := &models.Pvz{
		ID:               uuid.New(),
		City:             models.Moscow,
		RegistrationDate: time.Now(),
	}

	err := pvzService.CreatePvz(ctx, pvz)
	require.NoError(t, err, "Failed to create PVZ")

	reception := &models.Reception{
		// ID:       uuid.New(),
		// DateTime: time.Now(),
		PVZID:    pvz.ID,
		// Status:   models.ReceptionStatusInProgress,
	}

	_, err = receptionService.CreateReception(ctx, reception.PVZID)
	require.NoError(t, err, "Failed to create reception")

	products := make([]*models.Product, 50)
	for i := 0; i < 50; i++ {
		productType := models.ProductTypeClothing
		if i%2 == 0 {
			productType = models.ProductTypeElectronics
		}

		product, err := productService.AddProduct(ctx, productType, pvz.ID)
		require.NoError(t, err, "Failed to add product")
		products[i] = product
	}

	receptions, err := pvzService.GetPVZs(ctx, models.PvzFilter{
		Page:  1,
		Limit: 10,
	})
	require.NoError(t, err, "Failed to get PVZs with receptions")
	require.Len(t, receptions, 1)
	require.Len(t, receptions[0].Receptions, 1)
	assert.Len(t, receptions[0].Receptions[0].Products, 50)

	_, err = receptionService.CloseActiveReception(ctx, reception.PVZID)
	require.NoError(t, err, "Failed to close reception")

	updatedReceptions, err := pvzService.GetPVZs(ctx, models.PvzFilter{
		Page:  1,
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, models.ReceptionStatusClosed, updatedReceptions[0].Receptions[0].Reception.Status)

	_, err = productService.AddProduct(ctx, models.ProductTypeClothing, pvz.ID)
	assert.Error(t, err)
	assert.Equal(t, "no active reception", err.Error())

	err = productService.DeleteLastProduct(ctx, pvz.ID)
	assert.Error(t, err)
	assert.Equal(t, "no active reception", err.Error())

	newReception := &models.Reception{
		// ID:       uuid.New(),
		// DateTime: time.Now(),
		PVZID:    pvz.ID,
		// Status:   models.ReceptionStatusInProgress,
	}

	_, err = receptionService.CreateReception(ctx, newReception.PVZID)
	require.NoError(t, err)

	_, err = productService.AddProduct(ctx, models.ProductTypeClothing, pvz.ID)
	require.NoError(t, err)

	err = productService.DeleteLastProduct(ctx, pvz.ID)
	assert.NoError(t, err)
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	err = applyMigrations(pool)
	if err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	return pool
}

func applyMigrations(pool *pgxpool.Pool) error {

	_, err := pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS pvz (
			id UUID PRIMARY KEY,
			city TEXT NOT NULL,
			registration_date TIMESTAMP NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS receptions (
			id UUID PRIMARY KEY,
			date_time TIMESTAMP NOT NULL,
			status TEXT NOT NULL,
			pvz_id UUID REFERENCES pvz(id)
		);
		
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			date_time TIMESTAMP NOT NULL,
			type TEXT NOT NULL,
			reception_id UUID REFERENCES receptions(id)
		);
	`)
	return err
}
