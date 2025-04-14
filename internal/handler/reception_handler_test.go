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

	"avito-pvz/internal/dto"
	"avito-pvz/internal/mocks"
	"avito-pvz/internal/models"
)

func TestReceptionHandler_CreateReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := zerolog.New(nil)

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*mocks.MockReceptionService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "successful reception creation",
			requestBody: `{"pvzId": "f47ac10b-58cc-4372-a567-0e02b2c3d479"}`,
			mockSetup: func(ms *mocks.MockReceptionService) {
				ms.EXPECT().CreateReception(gomock.Any(), gomock.Any()).
					Return(&models.Reception{
						ID:       uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
						DateTime: time.Now(),
						PVZID:    uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
						Status:   models.ReceptionStatusInProgress,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: dto.Reception{
				Id:       uuidPtr("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
				DateTime: time.Now().Truncate(time.Second), // Truncate to avoid microsecond differences
				PvzId:    uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
				Status:   dto.InProgress,
			},
		},
		{
			name:           "invalid request body",
			requestBody:    `{"pvzId": "invalid-uuid"}`,
			mockSetup:      func(ms *mocks.MockReceptionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.Error{Message: "Invalid request"},
		},
		{
			name:        "service error",
			requestBody: `{"pvzId": "f47ac10b-58cc-4372-a567-0e02b2c3d479"}`,
			mockSetup: func(ms *mocks.MockReceptionService) {
				ms.EXPECT().CreateReception(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.Error{Message: "service error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockReceptionService(ctrl)
			handler := NewReceptionHandler(mockService, &logger)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(
				"POST",
				"/receptions",
				strings.NewReader(tt.requestBody),
			)
			ctx.Request.Header.Set("Content-Type", "application/json")

			tt.mockSetup(mockService)

			handler.CreateReception(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

			if tt.expectedStatus == http.StatusCreated {
				var response dto.Reception
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				expected := tt.expectedBody.(dto.Reception)
				assert.Equal(t, *expected.Id, *response.Id)
				assert.Equal(t, expected.PvzId, response.PvzId)
				assert.Equal(t, expected.Status, response.Status)
				assert.WithinDuration(t, expected.DateTime, response.DateTime, time.Second)
			} else {
				var response dto.Error
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(dto.Error).Message, response.Message)
			}
		})
	}
}

func TestReceptionHandler_CloseActiveReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := zerolog.New(nil)

	tests := []struct {
		name           string
		pvzID          string
		mockSetup      func(*mocks.MockReceptionService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:  "successful reception closure",
			pvzID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			mockSetup: func(ms *mocks.MockReceptionService) {
				ms.EXPECT().CloseActiveReception(gomock.Any(), gomock.Any()).
					Return(&models.Reception{
						ID:       uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
						DateTime: time.Now(),
						PVZID:    uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
						Status:   models.ReceptionStatusClosed,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: dto.Reception{
				Id:       uuidPtr("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
				DateTime: time.Now().Truncate(time.Second),
				PvzId:    uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
				Status:   dto.Close,
			},
		},
		{
			name:           "invalid pvz ID format",
			pvzID:          "invalid-uuid",
			mockSetup:      func(ms *mocks.MockReceptionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.Error{Message: "Invalid PVZ ID"},
		},
		{
			name:  "no active receptions",
			pvzID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			mockSetup: func(ms *mocks.MockReceptionService) {
				ms.EXPECT().CloseActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("no active receptions"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.Error{Message: "no active receptions"},
		},
		{
			name:  "service error",
			pvzID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			mockSetup: func(ms *mocks.MockReceptionService) {
				ms.EXPECT().CloseActiveReception(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   dto.Error{Message: "service error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockReceptionService(ctrl)
			handler := NewReceptionHandler(mockService, &logger)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(
				"POST",
				"/receptions/"+tt.pvzID,
				nil,
			)
			ctx.Params = gin.Params{
				{Key: "pvzId", Value: tt.pvzID},
			}

			tt.mockSetup(mockService)

			handler.CloseActiveReception(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

			if tt.expectedStatus == http.StatusOK {
				var response dto.Reception
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				expected := tt.expectedBody.(dto.Reception)
				assert.Equal(t, *expected.Id, *response.Id)
				assert.Equal(t, expected.PvzId, response.PvzId)
				assert.Equal(t, expected.Status, response.Status)
				assert.WithinDuration(t, expected.DateTime, response.DateTime, time.Second)
			} else {
				var response dto.Error
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(dto.Error).Message, response.Message)
			}
		})
	}
}

func uuidPtr(s string) *uuid.UUID {
	id := uuid.MustParse(s)
	return &id
}
