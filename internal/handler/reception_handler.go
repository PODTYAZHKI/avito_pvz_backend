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

type ReceptionService interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
}

type ReceptionHandler interface {
	CreateReception(c *gin.Context)
	CloseActiveReception(c *gin.Context)
}

type receptionHandler struct {
	receptionService ReceptionService
	logger           *zerolog.Logger
}

func NewReceptionHandler(receptionService ReceptionService, logger *zerolog.Logger) ReceptionHandler {
	return &receptionHandler{
		receptionService: receptionService,
		logger:           logger,
	}
}

func (h *receptionHandler) CreateReception(c *gin.Context) {
	var req dto.PostReceptionsJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind request")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid request"})
		return
	}

	reception, err := h.receptionService.CreateReception(c.Request.Context(), req.PvzId)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create reception")
		c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		return
	}

	response := dto.Reception{
		Id:       &reception.ID,
		DateTime: time.Time(reception.DateTime),
		PvzId:    reception.PVZID,
		Status:   dto.ReceptionStatus(reception.Status),
	}

	c.JSON(http.StatusCreated, response)
}

func (h *receptionHandler) CloseActiveReception(c *gin.Context) {
	pvzID, err := uuid.Parse(c.Param("pvzId"))
	if err != nil {
		h.logger.Error().Err(err).Msg("Invalid PVZ ID format")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid PVZ ID"})
		return
	}

	reception, err := h.receptionService.CloseActiveReception(c.Request.Context(), pvzID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "no active receptions" {
			status = http.StatusBadRequest
		}
		h.logger.Error().Err(err).Msg("Failed to close reception")
		c.JSON(status, dto.Error{Message: err.Error()})
		return
	}

	response := dto.Reception{
		Id:       &reception.ID,
		DateTime: time.Time(reception.DateTime),
		PvzId:    reception.PVZID,
		Status:   dto.ReceptionStatus(reception.Status),
	}

	c.JSON(http.StatusOK, response)
}
