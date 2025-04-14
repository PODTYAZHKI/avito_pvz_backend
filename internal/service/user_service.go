package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"avito-pvz/internal/config"
	"avito-pvz/internal/models"
	"avito-pvz/internal/utils"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserService interface {
	DummyLogin(role string) (*models.TokenResponse, error)
	Register(ctx context.Context, email, password, role string) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.TokenResponse, error)
}

type userService struct {
	userRepo UserRepository
	cfg      *config.Config
	logger   *zerolog.Logger
}

func NewUserService(repo UserRepository, cfg *config.Config, logger *zerolog.Logger) UserService {
	return &userService{
		userRepo: repo,
		cfg:      cfg,
		logger:   logger}
}

func (s *userService) DummyLogin(role string) (*models.TokenResponse, error) {
	if role != "employee" && role != "moderator" {
		return nil, models.ErrInvalidRole
	}
	token, err := utils.GenerateToken(s.cfg.SecretKey, role)
	if err != nil {
		return nil, models.ErrTokenGenerationFailed
	}
	return &models.TokenResponse{Token: token}, nil
}

func (s *userService) Register(ctx context.Context, email, password, role string) (*models.User, error) {
	if role != "employee" && role != "moderator" {
		return nil, errors.New("invalid role")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	user.Password = password

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*models.TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	token, err := utils.GenerateToken(s.cfg.SecretKey, user.Role)
	if err != nil {
		return nil, models.ErrTokenGenerationFailed
	}
	return &models.TokenResponse{Token: token}, nil
}
