package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"avito-pvz/internal/config"
	"avito-pvz/internal/mocks"
	"avito-pvz/internal/models"
	"avito-pvz/internal/utils"
)

func TestUserService_DummyLogin(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantErr bool
		err     error
	}{
		{
			name:    "success employee",
			role:    "employee",
			wantErr: false,
		},
		{
			name:    "success moderator",
			role:    "moderator",
			wantErr: false,
		},
		{
			name:    "invalid role",
			role:    "admin",
			wantErr: true,
			err:     models.ErrInvalidRole,
		},
	}

	cfg := &config.Config{SecretKey: "test-secret-key"}
	logger := zerolog.Nop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			service := NewUserService(repo, cfg, &logger)

			resp, err := service.DummyLogin(tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.err, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Token)

				claims, err := utils.ParseToken(resp.Token, cfg.SecretKey)
				assert.NoError(t, err)
				assert.Equal(t, tt.role, claims.Role)
			}
		})
	}
}

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		role      string
		mockSetup func(*mocks.MockUserRepository)
		wantErr   bool
		err       error
	}{
		{
			name:     "success employee",
			email:    "test@example.com",
			password: "password123",
			role:     "employee",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, user *models.User) {
						assert.Equal(t, "test@example.com", user.Email)
						assert.NotEqual(t, "password123", user.Password)
						assert.Equal(t, "employee", user.Role)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "invalid role",
			email:    "test@example.com",
			password: "password123",
			role:     "admin",
			mockSetup: func(m *mocks.MockUserRepository) {

			},
			wantErr: true,
			err:     errors.New("invalid role"),
		},
		{
			name:     "repository error",
			email:    "test@example.com",
			password: "password123",
			role:     "employee",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	cfg := &config.Config{}
	logger := zerolog.Nop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(repo)
			}

			service := NewUserService(repo, cfg, &logger)
			user, err := service.Register(context.Background(), tt.email, tt.password, tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.err != nil {
					assert.Equal(t, tt.err, err)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.password, user.Password)
				assert.Equal(t, tt.role, user.Role)
				assert.NotEqual(t, uuid.Nil, user.ID)
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	validPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		email     string
		password  string
		mockUser  *models.User
		mockError error
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "success",
			email:    "test@example.com",
			password: validPassword,
			mockUser: &models.User{
				Email:    "test@example.com",
				Password: string(hashedPassword),
				Role:     "employee",
			},
			wantErr: false,
		},
		{
			name:      "user not found",
			email:     "notfound@example.com",
			password:  validPassword,
			mockUser:  nil,
			mockError: nil,
			wantErr:   true,
			errMsg:    "user not found",
		},
		{
			name:     "invalid password",
			email:    "test@example.com",
			password: "wrongpassword",
			mockUser: &models.User{
				Email:    "test@example.com",
				Password: string(hashedPassword),
				Role:     "employee",
			},
			wantErr: true,
			errMsg:  "invalid password",
		},
		{
			name:      "repository error",
			email:     "test@example.com",
			password:  validPassword,
			mockUser:  nil,
			mockError: errors.New("db error"),
			wantErr:   true,
		},
	}

	cfg := &config.Config{SecretKey: "test-secret-key"}
	logger := zerolog.Nop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().
				FindByEmail(gomock.Any(), tt.email).
				Return(tt.mockUser, tt.mockError)

			service := NewUserService(repo, cfg, &logger)
			resp, err := service.Login(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotEmpty(t, resp.Token)

				claims, err := utils.ParseToken(resp.Token, cfg.SecretKey)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockUser.Role, claims.Role)
			}
		})
	}
}
