package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"avito-pvz/internal/mocks"
	"avito-pvz/internal/models"
)

func TestPvzService_CreatePvz(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPvzRepository(ctrl)
	logger := zerolog.New(nil)

	service := NewPvzService(mockRepo, &logger)

	tests := []struct {
		name        string
		input       *models.Pvz
		mockSetup   func()
		expectedErr error
	}{
		{
			name: "successful creation with all fields",
			input: &models.Pvz{
				City:             "Москва",
				RegistrationDate: time.Now().UTC().Add(-24 * time.Hour),
			},
			mockSetup: func() {
				mockRepo.EXPECT().CreatePvz(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid city",
			input: &models.Pvz{
				City: "INVALID_CITY",
			},
			mockSetup:   func() {},
			expectedErr: models.ErrInvalidCity,
		},
		{
			name: "repository error",
			input: &models.Pvz{
				City: "Москва",
			},
			mockSetup: func() {
				mockRepo.EXPECT().CreatePvz(gomock.Any(), gomock.Any()).Return(errors.New("repository error"))
			},
			expectedErr: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.CreatePvz(context.Background(), tt.input)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestPvzService_GetPVZs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPvzRepository(ctrl)
	logger := zerolog.New(nil)

	service := NewPvzService(mockRepo, &logger)

	now := time.Now().UTC()
	testPVZs := []models.PvzWithReceptions{
		{
			PVZ: models.Pvz{
				City:             "Москва",
				RegistrationDate: now.Add(-24 * time.Hour),
			},
			Receptions: []models.ReceptionWithProducts{},
		},
	}

	tests := []struct {
		name          string
		filter        models.PvzFilter
		mockSetup     func()
		expected      []models.PvzWithReceptions
		expectedErr   error
		expectedLimit int
		expectedPage  int
	}{
		{
			name: "successful get with default pagination",
			filter: models.PvzFilter{
				Page:  0,
				Limit: 0,
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetPVZsWithReceptions(gomock.Any(), models.PvzFilter{
					Page:  1,
					Limit: 10,
				}).Return(testPVZs, nil)
			},
			expected:    testPVZs,
			expectedErr: nil,
		},
		{
			name: "successful get with custom pagination",
			filter: models.PvzFilter{
				Page:  2,
				Limit: 5,
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetPVZsWithReceptions(gomock.Any(), models.PvzFilter{
					Page:  2,
					Limit: 5,
				}).Return(testPVZs[:1], nil)
			},
			expected:    testPVZs[:1],
			expectedErr: nil,
		},
		{
			name: "page less than 1",
			filter: models.PvzFilter{
				Page:  -1,
				Limit: 5,
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetPVZsWithReceptions(gomock.Any(), models.PvzFilter{
					Page:  1,
					Limit: 5,
				}).Return(testPVZs, nil)
			},
			expected:    testPVZs,
			expectedErr: nil,
		},
		{
			name: "limit less than 1",
			filter: models.PvzFilter{
				Page:  1,
				Limit: -1,
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetPVZsWithReceptions(gomock.Any(), models.PvzFilter{
					Page:  1,
					Limit: 10,
				}).Return(testPVZs, nil)
			},
			expected:    testPVZs,
			expectedErr: nil,
		},
		{
			name: "limit more than 30",
			filter: models.PvzFilter{
				Page:  1,
				Limit: 50,
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetPVZsWithReceptions(gomock.Any(), models.PvzFilter{
					Page:  1,
					Limit: 10,
				}).Return(testPVZs, nil)
			},
			expected:    testPVZs,
			expectedErr: nil,
		},
		{
			name: "repository error",
			filter: models.PvzFilter{
				Page:  1,
				Limit: 10,
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetPVZsWithReceptions(gomock.Any(), gomock.Any()).Return(nil, errors.New("repository error"))
			},
			expected:    nil,
			expectedErr: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.GetPVZs(context.Background(), tt.filter)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
