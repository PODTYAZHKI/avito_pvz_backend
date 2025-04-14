package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"avito-pvz/internal/mocks"
	"avito-pvz/internal/models"
)

func TestProductHandler_AddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockProductService(ctrl)
	logger := zerolog.New(nil)

	handler := NewProductHandler(mockService, &logger)

	fixedTime := time.Date(2025, 4, 13, 17, 37, 26, 496500000, time.FixedZone("MSK", 3*60*60))

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful product addition",
			requestBody: `{
                "type": "электроника",
                "pvzId": "550e8400-e29b-41d4-a716-446655440000"
            }`,
			mockSetup: func() {
				product := &models.Product{
					ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
					DateTime:    fixedTime,
					Type:        models.ProductTypeElectronics,
					ReceptionID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
				}
				mockService.EXPECT().AddProduct(
					gomock.Any(),
					models.ProductTypeElectronics,
					uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				).Return(product, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":          "550e8400-e29b-41d4-a716-446655440001",
				"dateTime":    fixedTime.Format(time.RFC3339Nano),
				"type":        "электроника",
				"receptionId": "550e8400-e29b-41d4-a716-446655440002",
			},
		},
		{
			name: "invalid request format",
			requestBody: `{
                "type": "электроника",
                "pvzId": "invalid-uuid"
            }`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "Invalid request format",
			},
		},
		{
			name: "invalid product type",
			requestBody: `{
                "type": "invalid-type",
                "pvzId": "550e8400-e29b-41d4-a716-446655440000"
            }`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "Invalid product type",
			},
		},
		{
			name: "no active reception",
			requestBody: `{
                "type": "электроника",
                "pvzId": "550e8400-e29b-41d4-a716-446655440000"
            }`,
			mockSetup: func() {
				mockService.EXPECT().AddProduct(
					gomock.Any(),
					models.ProductTypeElectronics,
					gomock.Any(),
				).Return(nil, errors.New("no active reception"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "no active reception",
			},
		},
		{
			name: "internal server error",
			requestBody: `{
                "type": "электроника",
                "pvzId": "550e8400-e29b-41d4-a716-446655440000"
            }`,
			mockSetup: func() {
				mockService.EXPECT().AddProduct(
					gomock.Any(),
					models.ProductTypeElectronics,
					gomock.Any(),
				).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"message": "internal error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(
				"POST",
				"/products",
				strings.NewReader(tt.requestBody),
			)
			ctx.Request.Header.Set("Content-Type", "application/json")

			tt.mockSetup()

			handler.AddProduct(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var actualBody map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				assert.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					if key == "dateTime" {

						_, err := time.Parse(time.RFC3339Nano, actualBody[key].(string))
						assert.NoError(t, err)
					} else {
						assert.Equal(t, expectedValue, actualBody[key], "field: %s", key)
					}
				}
			}
		})
	}
}

func TestProductHandler_DeleteLastProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockProductService(ctrl)
	logger := zerolog.New(nil)

	handler := NewProductHandler(mockService, &logger)

	tests := []struct {
		name           string
		pvzID          string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "successful deletion",
			pvzID: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func() {
				mockService.EXPECT().DeleteLastProduct(
					gomock.Any(),
					uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "invalid PVZ ID format",
			pvzID:          "invalid-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"message": "Invalid PVZ ID"
			}`,
		},
		{
			name:  "no active reception",
			pvzID: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func() {
				mockService.EXPECT().DeleteLastProduct(
					gomock.Any(),
					gomock.Any(),
				).Return(errors.New("no active reception"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"message": "no active reception"
			}`,
		},
		{
			name:  "no products to delete",
			pvzID: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func() {
				mockService.EXPECT().DeleteLastProduct(
					gomock.Any(),
					gomock.Any(),
				).Return(errors.New("no products to delete"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: `{
				"message": "no products to delete"
			}`,
		},
		{
			name:  "internal server error",
			pvzID: "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func() {
				mockService.EXPECT().DeleteLastProduct(
					gomock.Any(),
					gomock.Any(),
				).Return(errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: `{
				"message": "internal error"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(
				"DELETE",
				"/pvz/"+tt.pvzID+"/products/last",
				nil,
			)
			ctx.Params = gin.Params{
				{Key: "pvzId", Value: tt.pvzID},
			}

			tt.mockSetup()

			handler.DeleteLastProduct(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			} else {
				assert.Empty(t, w.Body.String())
			}
		})
	}
}
