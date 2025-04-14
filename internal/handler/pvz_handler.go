package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog"

	"avito-pvz/internal/dto"
	"avito-pvz/internal/models"
)

type PvzService interface {
	CreatePvz(ctx context.Context, pvz *models.Pvz) error
	GetPVZs(ctx context.Context, filter models.PvzFilter) ([]models.PvzWithReceptions, error)
}

type PvzHandler interface {
	CreatePvz(c *gin.Context)
	GetPvzs(c *gin.Context)
}

type pvzHandler struct {
	pvzService PvzService
	logger     *zerolog.Logger
}

func NewPvzHandler(pvzService PvzService, logger *zerolog.Logger) PvzHandler {
	return &pvzHandler{
		pvzService: pvzService,
		logger:     logger,
	}
}

func (h *pvzHandler) CreatePvz(c *gin.Context) {
	var pvz models.Pvz
	if err := c.ShouldBindJSON(&pvz); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind Pvz data")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid request"})
		return
	}



	if err := h.pvzService.CreatePvz(c.Request.Context(), &pvz); err != nil {
		h.logger.Error().Err(err).Msg("Failed to create Pvz")
		c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		return
	}
	response := dto.PVZ{
		City: dto.PVZCity(pvz.City),
		Id: (*types.UUID)(&pvz.ID),
		RegistrationDate: &pvz.RegistrationDate,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *pvzHandler) GetPvzs(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var startDate, endDate time.Time
	if startDateStr := c.Query("startDate"); startDateStr != "" {
		startDate, _ = time.Parse(time.RFC3339, startDateStr)
	}
	if endDateStr := c.Query("endDate"); endDateStr != "" {
		endDate, _ = time.Parse(time.RFC3339, endDateStr)
	}

	filter := models.PvzFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Limit:     limit,
	}

	results, err := h.pvzService.GetPVZs(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get PVZs")
		c.JSON(http.StatusInternalServerError, dto.Error{Message: "Failed to get PVZs"})
		return
	}

	type PvzResponse struct {
		PVZ        dto.PVZ `json:"pvz"`
		Receptions []struct {
			Reception dto.Reception `json:"reception"`
			Products  []dto.Product `json:"products"`
		} `json:"receptions"`
	}

	response := make([]PvzResponse, len(results))

	for i, item := range results {
		response[i].PVZ = dto.PVZ{
			Id:               &item.PVZ.ID,
			City:             dto.PVZCity(item.PVZ.City),
			RegistrationDate: &item.PVZ.RegistrationDate,
		}

		response[i].Receptions = make([]struct {
			Reception dto.Reception `json:"reception"`
			Products  []dto.Product `json:"products"`
		}, len(item.Receptions))

		for j, reception := range item.Receptions {
			response[i].Receptions[j].Reception = dto.Reception{
				Id:       &reception.Reception.ID,
				DateTime: reception.Reception.DateTime,
				PvzId:    reception.Reception.PVZID,
				Status:   dto.ReceptionStatus(reception.Reception.Status),
			}

			response[i].Receptions[j].Products = make([]dto.Product, len(reception.Products))
			for k, product := range reception.Products {
				response[i].Receptions[j].Products[k] = dto.Product{
					Id:          &product.ID,
					DateTime:    product.DateTime,
					Type:        dto.ProductType(product.Type),
					ReceptionId: product.ReceptionID,
				}
			}
		}
	}

	c.JSON(http.StatusOK, response)
}
