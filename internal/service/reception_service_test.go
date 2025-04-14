package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"avito-pvz/internal/mocks"
	"avito-pvz/internal/models"
)

func TestReceptionService_CreateReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReceptionRepository(ctrl)
	logger := zerolog.New(nil)

	service := NewReceptionService(mockRepo, &logger)

	tests := []struct {
		name           string
		pvzID          uuid.UUID
		mockSetup      func()
		expectedResult *models.Reception
		expectedError  error
	}{
		{
			name:  "successful reception creation",
			pvzID: uuid.New(),
			mockSetup: func() {

				mockRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, nil)

				mockRepo.EXPECT().
					CreateReception(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, r *models.Reception) error {
						assert.NotEqual(t, uuid.Nil, r.ID)
						assert.False(t, r.DateTime.IsZero())
						assert.Equal(t, models.ReceptionStatusInProgress, r.Status)
						return nil
					})
			},
			expectedResult: &models.Reception{
				Status: models.ReceptionStatusInProgress,
			},
			expectedError: nil,
		},
		{
			name:  "active reception already exists",
			pvzID: uuid.New(),
			mockSetup: func() {
				activeReception := &models.Reception{
					ID:     uuid.New(),
					Status: models.ReceptionStatusInProgress,
				}
				mockRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(activeReception, nil)
			},
			expectedResult: nil,
			expectedError:  errors.New("active reception already exists"),
		},
		{
			name:  "error checking active reception",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("repository error"),
		},
		{
			name:  "error creating reception",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, nil)
				mockRepo.EXPECT().
					CreateReception(gomock.Any(), gomock.Any()).
					Return(errors.New("repository error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.CreateReception(context.Background(), tt.pvzID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
				assert.NotEqual(t, uuid.Nil, result.ID)
				assert.False(t, result.DateTime.IsZero())
			}
		})
	}
}

func TestReceptionService_CloseActiveReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReceptionRepository(ctrl)
	logger := zerolog.New(nil)

	service := NewReceptionService(mockRepo, &logger)

	tests := []struct {
		name           string
		pvzID          uuid.UUID
		mockSetup      func()
		expectedResult *models.Reception
		expectedError  error
	}{
		{
			name:  "successful reception closing",
			pvzID: uuid.New(),
			mockSetup: func() {
				activeReception := &models.Reception{
					ID:     uuid.New(),
					Status: models.ReceptionStatusInProgress,
				}
				mockRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(activeReception, nil)
				mockRepo.EXPECT().
					CloseReception(gomock.Any(), activeReception.ID).
					Return(nil)
			},
			expectedResult: &models.Reception{
				Status: models.ReceptionStatusClosed,
			},
			expectedError: nil,
		},
		{
			name:  "error getting active reception",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("failed to get reception"),
		},
		{
			name:  "no active reception found",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  errors.New("no active receptions"),
		},
		{
			name:  "error closing reception",
			pvzID: uuid.New(),
			mockSetup: func() {
				activeReception := &models.Reception{
					ID:     uuid.New(),
					Status: models.ReceptionStatusInProgress,
				}
				mockRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(activeReception, nil)
				mockRepo.EXPECT().
					CloseReception(gomock.Any(), activeReception.ID).
					Return(errors.New("repository error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("failed to close reception"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := service.CloseActiveReception(context.Background(), tt.pvzID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
			}
		})
	}
}
