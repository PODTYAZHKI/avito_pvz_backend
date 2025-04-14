package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"avito-pvz/internal/metrics"
	"avito-pvz/internal/models"
)

type PvzRepository interface {
	CreatePvz(ctx context.Context, pvz *models.Pvz) error
	GetPVZsWithReceptions(ctx context.Context, filter models.PvzFilter) ([]models.PvzWithReceptions, error)
}

type PvzService interface {
	CreatePvz(ctx context.Context, pvz *models.Pvz) error
	GetPVZs(ctx context.Context, filter models.PvzFilter) ([]models.PvzWithReceptions, error)
}

type pvzService struct {
	pvzRepo PvzRepository
	logger  *zerolog.Logger
}

func NewPvzService(repo PvzRepository, logger *zerolog.Logger) PvzService {
	return &pvzService{
		pvzRepo: repo,
		logger:  logger,
	}
}

func (s *pvzService) CreatePvz(ctx context.Context, pvz *models.Pvz) error {
	if !pvz.City.IsValid() {
		s.logger.Warn().Str("city", string(pvz.City)).Msg("Invalid city provided")
		return models.ErrInvalidCity
	}

	if pvz.ID == uuid.Nil {
		pvz.ID = uuid.New()
	}

	if pvz.RegistrationDate.IsZero() {
		pvz.RegistrationDate = time.Now().UTC()
	}
	metrics.PvzCreatedTotal.Inc()
	return s.pvzRepo.CreatePvz(ctx, pvz)
}

func (s *pvzService) GetPVZs(ctx context.Context, filter models.PvzFilter) ([]models.PvzWithReceptions, error) {

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 30 {
		filter.Limit = 10
	}

	results, err := s.pvzRepo.GetPVZsWithReceptions(ctx, filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get PVZs with receptions")
		return nil, err
	}

	s.logger.Info().Int("count", len(results)).Msg("Successfully retrieved PVZs")
	return results, nil
}
