package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"avito-pvz/internal/models"
)

type PvzRepository interface {
	CreatePvz(ctx context.Context, pvz *models.Pvz) error
	GetPVZsWithReceptions(ctx context.Context, filter models.PvzFilter) ([]models.PvzWithReceptions, error)
}

type pvzRepository struct {
	pool    PgxPool
	builder squirrel.StatementBuilderType
	logger  *zerolog.Logger
}

func NewPvzRepository(pool PgxPool, logger *zerolog.Logger) PvzRepository {
	return &pvzRepository{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger,
	}
}

func (r *pvzRepository) CreatePvz(ctx context.Context, pvz *models.Pvz) error {
	query, args, err := r.builder.Insert("pvz").
		Columns("id", "registration_date", "city").
		Values(pvz.ID, pvz.RegistrationDate, pvz.City).
		ToSql()
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to build SQL query")
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("city", string(pvz.City)).Msg("Failed to create Pvz")
		return err
	}

	return nil
}

func (r *pvzRepository) GetPVZsWithReceptions(ctx context.Context, filter models.PvzFilter) ([]models.PvzWithReceptions, error) {
	pvzs, err := r.getPaginatedPVZs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get paginated PVZs failed: %w", err)
	}
	if len(pvzs) == 0 {
		return []models.PvzWithReceptions{}, nil
	}

	receptionsWithProducts, err := r.getAllReceptionsWithProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("get receptions for PVZs failed: %w", err)
	}

	return r.buildFinalResult(pvzs, receptionsWithProducts), nil
}

func (r *pvzRepository) getPaginatedPVZs(ctx context.Context, filter models.PvzFilter) (map[uuid.UUID]models.Pvz, error) {
	query := r.builder.
		Select("id", "city", "registration_date").
		From("pvz").
		Offset(uint64((filter.Page - 1) * filter.Limit)).
		Limit(uint64(filter.Limit))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]models.Pvz)
	for rows.Next() {
		var pvz models.Pvz
		if err := rows.Scan(&pvz.ID, &pvz.City, &pvz.RegistrationDate); err != nil {
			return nil, err
		}
		result[pvz.ID] = pvz
	}

	return result, nil
}

func (r *pvzRepository) getAllReceptionsWithProducts(ctx context.Context) (map[uuid.UUID][]models.ReceptionWithProducts, error) {

	receptionsQuery := r.builder.
		Select("id", "date_time", "status", "pvz_id").
		From("receptions").
		OrderBy("pvz_id, date_time DESC")

	receptionsSQL, _, err := receptionsQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build receptions query: %w", err)
	}

	receptionsRows, err := r.pool.Query(ctx, receptionsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query receptions: %w", err)
	}
	defer receptionsRows.Close()

	result := make(map[uuid.UUID][]models.ReceptionWithProducts)
	var receptionIDs []uuid.UUID

	for receptionsRows.Next() {
		var reception models.Reception
		if err := receptionsRows.Scan(
			&reception.ID,
			&reception.DateTime,
			&reception.Status,
			&reception.PVZID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan reception: %w", err)
		}

		result[reception.PVZID] = append(result[reception.PVZID], models.ReceptionWithProducts{
			Reception: reception,
			Products:  []models.Product{},
		})
		receptionIDs = append(receptionIDs, reception.ID)
	}

	if len(receptionIDs) == 0 {
		return result, nil
	}

	productsQuery := r.builder.
		Select("id", "date_time", "type", "reception_id").
		From("products").
		OrderBy("reception_id, date_time DESC")

	productsSQL, _, err := productsQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build products query: %w", err)
	}

	productsRows, err := r.pool.Query(ctx, productsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer productsRows.Close()

	receptionsMap := make(map[uuid.UUID]*models.ReceptionWithProducts)
	for pvzID, receptions := range result {
		for i := range receptions {
			receptionsMap[receptions[i].Reception.ID] = &result[pvzID][i]
		}
	}

	for productsRows.Next() {
		var product models.Product
		if err := productsRows.Scan(
			&product.ID,
			&product.DateTime,
			&product.Type,
			&product.ReceptionID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		if reception, exists := receptionsMap[product.ReceptionID]; exists {
			reception.Products = append(reception.Products, product)
		}
	}

	return result, nil
}

func (r *pvzRepository) buildFinalResult(pvzs map[uuid.UUID]models.Pvz, receptions map[uuid.UUID][]models.ReceptionWithProducts) []models.PvzWithReceptions {
	result := make([]models.PvzWithReceptions, 0, len(pvzs))

	for pvzID, pvz := range pvzs {
		pvzReceptions, exists := receptions[pvzID]
		if !exists {
			pvzReceptions = []models.ReceptionWithProducts{}
		}

		result = append(result, models.PvzWithReceptions{
			PVZ:        pvz,
			Receptions: pvzReceptions,
		})
	}

	return result
}
