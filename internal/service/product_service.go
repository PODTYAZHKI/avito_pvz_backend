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

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetLastProduct(ctx context.Context, pvzID uuid.UUID) (*models.Product, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
}

type ProductService interface {
	AddProduct(ctx context.Context, productType models.ProductType, pvzID uuid.UUID) (*models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

type productService struct {
	receptionRepo ReceptionRepository
	productRepo   ProductRepository
	logger        *zerolog.Logger
}

func NewProductService(
	receptionRepo ReceptionRepository,
	productRepo ProductRepository,
	logger *zerolog.Logger,
) ProductService {
	return &productService{
		receptionRepo: receptionRepo,
		productRepo:   productRepo,
		logger:        logger,
	}
}

func (s *productService) AddProduct(
	ctx context.Context,
	productType models.ProductType,
	pvzID uuid.UUID,
) (*models.Product, error) {
	if !productType.IsValid() {
		s.logger.Warn().Str("type", string(productType)).Msg("Invalid product type")
		return nil, models.ErrInvalidProductType
	}

	reception, err := s.receptionRepo.GetActiveReception(ctx, pvzID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("pvzID", pvzID.String()).
			Msg("Failed to get active reception")
		return nil, errors.New("failed to check active reception")
	}

	if reception == nil {
		s.logger.Warn().
			Str("pvzID", pvzID.String()).
			Msg("No active reception found")
		return nil, models.ErrNoActiveReception
	}

	product := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now().UTC(),
		Type:        productType,
		ReceptionID: reception.ID,
	}

	if err := s.productRepo.CreateProduct(ctx, product); err != nil {
		s.logger.Error().
			Err(err).
			Str("productID", product.ID.String()).
			Msg("Failed to create product")
		return nil, errors.New("failed to create product")
	}

	s.logger.Info().
		Str("productID", product.ID.String()).
		Str("receptionID", reception.ID.String()).
		Str("pvzID", pvzID.String()).
		Msg("Product added successfully")

	metrics.ProductAddedTotal.Inc()
	return product, nil
}

func (s *productService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {

	reception, err := s.receptionRepo.GetActiveReception(ctx, pvzID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("pvzID", pvzID.String()).
			Msg("Failed to check active reception")
		return errors.New("failed to check active reception")
	}

	if reception == nil {
		s.logger.Warn().
			Str("pvzID", pvzID.String()).
			Msg("No active reception found")
		return models.ErrNoActiveReception
	}

	product, err := s.productRepo.GetLastProduct(ctx, pvzID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("pvzID", pvzID.String()).
			Msg("Failed to get last product")
		return errors.New("failed to get last product")
	}

	if product == nil {
		s.logger.Warn().
			Str("pvzID", pvzID.String()).
			Msg("No products to delete")
		return errors.New("no products to delete")
	}

	if err := s.productRepo.DeleteProduct(ctx, product.ID); err != nil {
		s.logger.Error().
			Err(err).
			Str("productID", product.ID.String()).
			Msg("Failed to delete product")
		return errors.New("failed to delete product")
	}

	s.logger.Info().
		Str("productID", product.ID.String()).
		Str("receptionID", reception.ID.String()).
		Str("pvzID", pvzID.String()).
		Msg("Product deleted successfully")

	return nil
}
