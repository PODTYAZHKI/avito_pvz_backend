package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"avito-pvz/internal/dto"
	"avito-pvz/internal/mocks"
	"avito-pvz/internal/models"
)

func TestUserHandler_DummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserService(ctrl)
	logger := zerolog.New(nil)
	h := NewUserHandler(mockUserService, &logger)

	tests := []struct {
		name           string
		role           string
		mockReturn     *models.TokenResponse
		mockError      error
		expectedStatus int
		expectedResult string
	}{
		{
			name:           "success",
			role:           "employee",
			mockReturn:     &models.TokenResponse{Token: "fake-token"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedResult: `{"token":"fake-token"}`,
		},
		{
			name:           "invalid role",
			role:           "invalid",
			mockReturn:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedResult: `{"message":"Invalid role"}`,
		},
		{
			name:           "service error",
			role:           "moderator",
			mockReturn:     nil,
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedResult: `{"message":"Failed to generate token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.role == "employee" || tt.role == "moderator" {
				mockUserService.EXPECT().
					DummyLogin(tt.role).
					Return(tt.mockReturn, tt.mockError).
					Times(1)
			}

			reqBody, err := json.Marshal(dto.PostDummyLoginJSONBody{Role: dto.PostDummyLoginJSONBodyRole(tt.role)})
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/dummy-login", bytes.NewBuffer(reqBody))
			ctx.Request.Header.Set("Content-Type", "application/json")

			ctx.Set("role", tt.role)

			h.DummyLogin(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedResult, w.Body.String())
		})
	}
}
