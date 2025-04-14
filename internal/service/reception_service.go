package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"avito-pvz/internal/metrics"
	"avito-pvz/internal/models"
)

type ReceptionRepository interface {
	CreateReception(ctx context.Context, reception *models.Reception) error
	GetActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, receptionID uuid.UUID) error
}

type ReceptionService interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
}

type receptionService struct {
	receptionRepo ReceptionRepository
	logger        *zerolog.Logger
}

func NewReceptionService(repo ReceptionRepository, logger *zerolog.Logger) ReceptionService {
	return &receptionService{
		receptionRepo: repo,
		logger:        logger,
	}
}

func (s *receptionService) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	activeReception, err := s.receptionRepo.GetActiveReception(ctx, pvzID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to check active receptions")
		return nil, err
	}

	if activeReception != nil {
		s.logger.Warn().Str("pvzID", pvzID.String()).Msg("Active reception already exists")
		return nil, errors.New("active reception already exists")
	}

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now().UTC(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	if err := s.receptionRepo.CreateReception(ctx, reception); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create reception")
		return nil, err
	}

	s.logger.Info().
		Str("receptionID", reception.ID.String()).
		Str("pvzID", pvzID.String()).
		Msg("Reception created successfully")

	metrics.ReceptionCreatedTotal.Inc()
	return reception, nil
}

func (s *receptionService) CloseActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {

	reception, err := s.receptionRepo.GetActiveReception(ctx, pvzID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("pvzID", pvzID.String()).
			Msg("Failed to get active reception")
		return nil, errors.New("failed to get reception")
	}

	if reception == nil {
		s.logger.Warn().
			Str("pvzID", pvzID.String()).
			Msg("No active receptions found")
		return nil, errors.New("no active receptions")
	}

	if err := s.receptionRepo.CloseReception(ctx, reception.ID); err != nil {
		s.logger.Error().
			Err(err).
			Str("receptionID", reception.ID.String()).
			Msg("Failed to close reception")
		return nil, errors.New("failed to close reception")
	}

	reception.Status = models.ReceptionStatusClosed

	s.logger.Info().
		Str("receptionID", reception.ID.String()).
		Str("pvzID", pvzID.String()).
		Msg("Reception closed successfully")

	return reception, nil
}
