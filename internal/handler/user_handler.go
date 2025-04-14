package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog"

	"avito-pvz/internal/dto"
	"avito-pvz/internal/models"
)

type UserService interface {
	DummyLogin(role string) (*models.TokenResponse, error)
	Register(ctx context.Context, email, password, role string) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.TokenResponse, error)
}

type UserHandler interface {
	DummyLogin(c *gin.Context)
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type userHandler struct {
	userService UserService
	logger      *zerolog.Logger
}

func NewUserHandler(userService UserService, logger *zerolog.Logger) UserHandler {
	return &userHandler{userService: userService, logger: logger}
}

func (h *userHandler) DummyLogin(c *gin.Context) {
	var req dto.PostDummyLoginJSONBody

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind request body")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid request body"})
		return
	}

	if req.Role != "employee" && req.Role != "moderator" {
		h.logger.Warn().Str("role", string(req.Role)).Msg("Validation failed")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid role"})
		return
	}

	tokenResponse, err := h.userService.DummyLogin(string(req.Role))
	if err != nil {
		h.logger.Error().Err(err).Str("role", string(req.Role)).Msg("Dummy login failed")
		c.JSON(http.StatusInternalServerError, dto.Error{Message: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, tokenResponse)
}

func (h *userHandler) Register(c *gin.Context) {
	var req dto.PostRegisterJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind request")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid request"})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), string(req.Email), req.Password, string(req.Role))
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create user")
		c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		return
	}

	response := dto.PostRegisterJSONBody{
		Email:    types.Email(user.Email),
		Password: user.Password,
		Role:     dto.PostRegisterJSONBodyRole(user.Role),
	}

	c.JSON(http.StatusCreated, response)
}

func (h *userHandler) Login(c *gin.Context) {
	var req dto.PostLoginJSONBody

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind request body")
		c.JSON(http.StatusBadRequest, dto.Error{Message: "Invalid request body"})
		return
	}

	tokenResponse, err := h.userService.Login(c.Request.Context(), string(req.Email), req.Password)
	if err != nil {
		h.logger.Error().Err(err).Msg("Login failed")
		c.JSON(http.StatusInternalServerError, dto.Error{Message: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, tokenResponse)
}
