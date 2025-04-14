package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"avito-pvz/internal/dto"
	"avito-pvz/internal/models"
)

type ProductService interface {
	AddProduct(ctx context.Context, productType models.ProductType, pvzID uuid.UUID) (*models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

type ProductHandler interface {
	AddProduct(c *gin.Context)
	DeleteLastProduct(c *gin.Context)
}

type productHandler struct {
	productService ProductService
	logger         *zerolog.Logger
}

func NewProductHandler(productService ProductService, logger *zerolog.Logger) ProductHandler {
	return &productHandler{
		productService: productService,
		logger:         logger,
	}
}

func (h *productHandler) AddProduct(c *gin.Context) {
	var req dto.PostProductsJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request format")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid request format"})
		return
	}

	productType := models.ProductType(req.Type)
	if !productType.IsValid() {
		h.logger.Warn().Str("type", string(req.Type)).Msg("Invalid product type")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid product type"})
		return
	}

	product, err := h.productService.AddProduct(c.Request.Context(), productType, req.PvzId)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "no active reception" {
			status = http.StatusBadRequest
		}
		h.logger.Error().Err(err).Msg("Failed to add product")
		c.JSON(status, dto.Error{Message: err.Error()})
		return
	}

	response := dto.Product{
		Id:          &product.ID,
		DateTime:    time.Time(product.DateTime),
		Type:        dto.ProductType(product.Type),
		ReceptionId: product.ReceptionID,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *productHandler) DeleteLastProduct(c *gin.Context) {
	pvzID, err := uuid.Parse(c.Param("pvzId"))
	if err != nil {
		h.logger.Error().Err(err).Msg("Invalid PVZ ID format")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid PVZ ID"})
		return
	}

	if err := h.productService.DeleteLastProduct(c.Request.Context(), pvzID); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "no active reception" || err.Error() == "no products to delete" {
			status = http.StatusBadRequest
		}
		h.logger.Error().Err(err).Msg("Failed to delete last product")
		c.JSON(status, dto.Error{Message: err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
