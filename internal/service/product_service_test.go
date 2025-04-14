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

func TestProductService_AddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockReceptionRepository(ctrl)
	mockProductRepo := mocks.NewMockProductRepository(ctrl)
	logger := zerolog.New(nil)

	service := NewProductService(mockReceptionRepo, mockProductRepo, &logger)

	testCases := []struct {
		name           string
		productType    models.ProductType
		pvzID          uuid.UUID
		mockSetup      func()
		expectedResult *models.Product
		expectedError  error
	}{
		{
			name:        "successful product addition",
			productType: models.ProductTypeElectronics,
			pvzID:       uuid.New(),
			mockSetup: func() {
				activeReception := &models.Reception{
					ID:     uuid.New(),
					PVZID:  uuid.New(),
					Status: models.ReceptionStatusInProgress,
				}
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(activeReception, nil)
				mockProductRepo.EXPECT().
					CreateProduct(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, product *models.Product) error {
						assert.NotEqual(t, uuid.Nil, product.ID)
						assert.False(t, product.DateTime.IsZero())
						assert.Equal(t, models.ProductTypeElectronics, product.Type)
						assert.Equal(t, activeReception.ID, product.ReceptionID)
						return nil
					})
			},
			expectedResult: &models.Product{
				Type: models.ProductTypeElectronics,
			},
			expectedError: nil,
		},
		{
			name:           "invalid product type",
			productType:    "invalid_type",
			pvzID:          uuid.New(),
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError:  models.ErrInvalidProductType,
		},
		{
			name:        "no active reception",
			productType: models.ProductTypeClothing,
			pvzID:       uuid.New(),
			mockSetup: func() {
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  models.ErrNoActiveReception,
		},
		{
			name:        "error getting active reception",
			productType: models.ProductTypeShoes,
			pvzID:       uuid.New(),
			mockSetup: func() {
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("failed to check active reception"),
		},
		{
			name:        "error creating product",
			productType: models.ProductTypeElectronics,
			pvzID:       uuid.New(),
			mockSetup: func() {
				activeReception := &models.Reception{
					ID:     uuid.New(),
					PVZID:  uuid.New(),
					Status: models.ReceptionStatusInProgress,
				}
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(activeReception, nil)
				mockProductRepo.EXPECT().
					CreateProduct(gomock.Any(), gomock.Any()).
					Return(errors.New("repository error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("failed to create product"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			result, err := service.AddProduct(context.Background(), tc.productType, tc.pvzID)

			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedResult.Type, result.Type)
				assert.NotEqual(t, uuid.Nil, result.ID)
				assert.False(t, result.DateTime.IsZero())
			}
		})
	}
}

func TestProductService_DeleteLastProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReceptionRepo := mocks.NewMockReceptionRepository(ctrl)
	mockProductRepo := mocks.NewMockProductRepository(ctrl)
	logger := zerolog.New(nil)

	service := NewProductService(mockReceptionRepo, mockProductRepo, &logger)

	testCases := []struct {
		name          string
		pvzID         uuid.UUID
		mockSetup     func()
		expectedError error
	}{
		{
			name:  "successful product deletion",
			pvzID: uuid.New(),
			mockSetup: func() {
				activeReception := &models.Reception{
					ID:     uuid.New(),
					PVZID:  uuid.New(),
					Status: models.ReceptionStatusInProgress,
				}
				lastProduct := &models.Product{
					ID:          uuid.New(),
					ReceptionID: activeReception.ID,
				}
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(activeReception, nil)
				mockProductRepo.EXPECT().
					GetLastProduct(gomock.Any(), gomock.Any()).
					Return(lastProduct, nil)
				mockProductRepo.EXPECT().
					DeleteProduct(gomock.Any(), lastProduct.ID).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "error getting active reception",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			expectedError: errors.New("failed to check active reception"),
		},
		{
			name:  "no active reception",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedError: models.ErrNoActiveReception,
		},
		{
			name:  "error getting last product",
			pvzID: uuid.New(),
			mockSetup: func() {
				activeReception := &models.Reception{
					ID:     uuid.New(),
					PVZID:  uuid.New(),
					Status: models.ReceptionStatusInProgress,
				}
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(activeReception, nil)
				mockProductRepo.EXPECT().
					GetLastProduct(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("repository error"))
			},
			expectedError: errors.New("failed to get last product"),
		},
		{
			name:  "no products to delete",
			pvzID: uuid.New(),
			mockSetup: func() {
				activeReception := &models.Reception{
					ID:     uuid.New(),
					PVZID:  uuid.New(),
					Status: models.ReceptionStatusInProgress,
				}
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(activeReception, nil)
				mockProductRepo.EXPECT().
					GetLastProduct(gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedError: errors.New("no products to delete"),
		},
		{
			name:  "error deleting product",
			pvzID: uuid.New(),
			mockSetup: func() {
				activeReception := &models.Reception{
					ID:     uuid.New(),
					PVZID:  uuid.New(),
					Status: models.ReceptionStatusInProgress,
				}
				lastProduct := &models.Product{
					ID:          uuid.New(),
					ReceptionID: activeReception.ID,
				}
				mockReceptionRepo.EXPECT().
					GetActiveReception(gomock.Any(), gomock.Any()).
					Return(activeReception, nil)
				mockProductRepo.EXPECT().
					GetLastProduct(gomock.Any(), gomock.Any()).
					Return(lastProduct, nil)
				mockProductRepo.EXPECT().
					DeleteProduct(gomock.Any(), lastProduct.ID).
					Return(errors.New("repository error"))
			},
			expectedError: errors.New("failed to delete product"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := service.DeleteLastProduct(context.Background(), tc.pvzID)

			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
