package handler

import (
	"context"
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

	"avito-pvz/internal/dto"
	"avito-pvz/internal/mocks"
	"avito-pvz/internal/models"
)

func TestPvzHandler_CreatePvz(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := zerolog.New(nil)

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*mocks.MockPvzService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "successful PVZ creation",
			requestBody: `{"city": "Москва"}`,
			mockSetup: func(ms *mocks.MockPvzService) {
				ms.EXPECT().CreatePvz(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, pvz *models.Pvz) error {
						pvz.ID = uuid.New()
						pvz.RegistrationDate = time.Now()
						return nil
					})
			},
			expectedStatus: http.StatusCreated,
			expectedBody: models.Pvz{
				City: "Москва",
			},
		},
		{
			name:           "invalid request body",
			requestBody:    `{"city": "Москва",}`,
			mockSetup:      func(ms *mocks.MockPvzService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.Error{Message: "Invalid request"},
		},
		{
			name:        "invalid city",
			requestBody: `{"city": "InvalidCity"}`,
			mockSetup: func(ms *mocks.MockPvzService) {
				ms.EXPECT().CreatePvz(gomock.Any(), gomock.Any()).
					Return(models.ErrInvalidCity)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.Error{Message: models.ErrInvalidCity.Error()},
		},
		{
			name:        "service error",
			requestBody: `{"city": "Москва"}`,
			mockSetup: func(ms *mocks.MockPvzService) {
				ms.EXPECT().CreatePvz(gomock.Any(), gomock.Any()).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.Error{Message: "service error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockPvzService(ctrl)
			handler := NewPvzHandler(mockService, &logger)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(
				"POST",
				"/pvz",
				strings.NewReader(tt.requestBody),
			)
			ctx.Request.Header.Set("Content-Type", "application/json")

			tt.mockSetup(mockService)

			handler.CreatePvz(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response models.Pvz
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, tt.expectedBody.(models.Pvz).City, response.City)
				assert.NotEqual(t, uuid.Nil, response.ID)
				assert.False(t, response.RegistrationDate.IsZero())
			} else {
				var response dto.Error
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(dto.Error).Message, response.Message)
			}
		})
	}
}

func TestPvzHandler_GetPvzs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockPvzService(ctrl)
	logger := zerolog.New(nil)

	handler := NewPvzHandler(mockService, &logger)

	now := time.Now().UTC()
	testPVZs := []models.PvzWithReceptions{
		{
			PVZ: models.Pvz{
				ID:               uuid.New(),
				City:             models.Moscow,
				RegistrationDate: now.Add(-24 * time.Hour),
			},
			Receptions: []models.ReceptionWithProducts{
				{
					Reception: models.Reception{
						ID:       uuid.New(),
						DateTime: now.Add(-12 * time.Hour),
						PVZID:    uuid.New(),
						Status:   models.ReceptionStatusInProgress,
					},
					Products: []models.Product{
						{
							ID:          uuid.New(),
							DateTime:    now.Add(-6 * time.Hour),
							Type:        models.ProductTypeElectronics,
							ReceptionID: uuid.New(),
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "successful get with default params",
			queryParams: map[string]string{},
			mockSetup: func() {
				mockService.EXPECT().GetPVZs(
					gomock.Any(),
					models.PvzFilter{
						Page:  1,
						Limit: 10,
					},
				).Return(testPVZs, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []struct {
				PVZ        dto.PVZ `json:"pvz"`
				Receptions []struct {
					Reception dto.Reception `json:"reception"`
					Products  []dto.Product `json:"products"`
				} `json:"receptions"`
			}{
				{
					PVZ: dto.PVZ{
						Id:               &testPVZs[0].PVZ.ID,
						City:             dto.PVZCity(testPVZs[0].PVZ.City),
						RegistrationDate: &testPVZs[0].PVZ.RegistrationDate,
					},
					Receptions: []struct {
						Reception dto.Reception `json:"reception"`
						Products  []dto.Product `json:"products"`
					}{
						{
							Reception: dto.Reception{
								Id:       &testPVZs[0].Receptions[0].Reception.ID,
								DateTime: testPVZs[0].Receptions[0].Reception.DateTime,
								PvzId:    testPVZs[0].Receptions[0].Reception.PVZID,
								Status:   dto.ReceptionStatus(testPVZs[0].Receptions[0].Reception.Status),
							},
							Products: []dto.Product{
								{
									Id:          &testPVZs[0].Receptions[0].Products[0].ID,
									DateTime:    testPVZs[0].Receptions[0].Products[0].DateTime,
									Type:        dto.ProductType(testPVZs[0].Receptions[0].Products[0].Type),
									ReceptionId: testPVZs[0].Receptions[0].Products[0].ReceptionID,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "with custom pagination",
			queryParams: map[string]string{
				"page":  "2",
				"limit": "5",
			},
			mockSetup: func() {
				mockService.EXPECT().GetPVZs(
					gomock.Any(),
					models.PvzFilter{
						Page:  2,
						Limit: 5,
					},
				).Return(testPVZs, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []struct {
				PVZ        dto.PVZ `json:"pvz"`
				Receptions []struct {
					Reception dto.Reception `json:"reception"`
					Products  []dto.Product `json:"products"`
				} `json:"receptions"`
			}{
				{
					PVZ: dto.PVZ{
						Id:               &testPVZs[0].PVZ.ID,
						City:             dto.PVZCity(testPVZs[0].PVZ.City),
						RegistrationDate: &testPVZs[0].PVZ.RegistrationDate,
					},
					Receptions: []struct {
						Reception dto.Reception `json:"reception"`
						Products  []dto.Product `json:"products"`
					}{
						{
							Reception: dto.Reception{
								Id:       &testPVZs[0].Receptions[0].Reception.ID,
								DateTime: testPVZs[0].Receptions[0].Reception.DateTime,
								PvzId:    testPVZs[0].Receptions[0].Reception.PVZID,
								Status:   dto.ReceptionStatus(testPVZs[0].Receptions[0].Reception.Status),
							},
							Products: []dto.Product{
								{
									Id:          &testPVZs[0].Receptions[0].Products[0].ID,
									DateTime:    testPVZs[0].Receptions[0].Products[0].DateTime,
									Type:        dto.ProductType(testPVZs[0].Receptions[0].Products[0].Type),
									ReceptionId: testPVZs[0].Receptions[0].Products[0].ReceptionID,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "with date filters",
			queryParams: map[string]string{
				"startDate": now.Add(-48 * time.Hour).Format(time.RFC3339),
				"endDate":   now.Format(time.RFC3339),
			},
			mockSetup: func() {
				mockService.EXPECT().GetPVZs(
					gomock.Any(),
					gomock.Any(),
				).DoAndReturn(func(ctx context.Context, filter models.PvzFilter) ([]models.PvzWithReceptions, error) {
					assert.False(t, filter.StartDate.IsZero())
					assert.False(t, filter.EndDate.IsZero())
					return testPVZs, nil
				})
			},
			expectedStatus: http.StatusOK,
			expectedBody: []struct {
				PVZ        dto.PVZ `json:"pvz"`
				Receptions []struct {
					Reception dto.Reception `json:"reception"`
					Products  []dto.Product `json:"products"`
				} `json:"receptions"`
			}{
				{
					PVZ: dto.PVZ{
						Id:               &testPVZs[0].PVZ.ID,
						City:             dto.PVZCity(testPVZs[0].PVZ.City),
						RegistrationDate: &testPVZs[0].PVZ.RegistrationDate,
					},
					Receptions: []struct {
						Reception dto.Reception `json:"reception"`
						Products  []dto.Product `json:"products"`
					}{
						{
							Reception: dto.Reception{
								Id:       &testPVZs[0].Receptions[0].Reception.ID,
								DateTime: testPVZs[0].Receptions[0].Reception.DateTime,
								PvzId:    testPVZs[0].Receptions[0].Reception.PVZID,
								Status:   dto.ReceptionStatus(testPVZs[0].Receptions[0].Reception.Status),
							},
							Products: []dto.Product{
								{
									Id:          &testPVZs[0].Receptions[0].Products[0].ID,
									DateTime:    testPVZs[0].Receptions[0].Products[0].DateTime,
									Type:        dto.ProductType(testPVZs[0].Receptions[0].Products[0].Type),
									ReceptionId: testPVZs[0].Receptions[0].Products[0].ReceptionID,
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "service error",
			queryParams: map[string]string{},
			mockSetup: func() {
				mockService.EXPECT().GetPVZs(
					gomock.Any(),
					gomock.Any(),
				).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   dto.Error{Message: "Failed to get PVZs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/pvz", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			ctx.Request = req

			tt.mockSetup()

			handler.GetPvzs(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []struct {
					PVZ        dto.PVZ `json:"pvz"`
					Receptions []struct {
						Reception dto.Reception `json:"reception"`
						Products  []dto.Product `json:"products"`
					} `json:"receptions"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			} else {
				var response dto.Error
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(dto.Error).Message, response.Message)
			}
		})
	}
}
