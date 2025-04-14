package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"avito-pvz/internal/config"
	"avito-pvz/internal/dto"
	"avito-pvz/internal/models"
	"avito-pvz/internal/utils"
)

func TestAuthMiddleware(t *testing.T) {

	gin.SetMode(gin.TestMode)

	secretKey := "test-secret-key"
	cfg := &config.Config{SecretKey: secretKey}
	logger := zerolog.New(nil)

	validModeratorToken, _ := utils.GenerateToken(secretKey, models.RoleModerator)
	validEmployeeToken, _ := utils.GenerateToken(secretKey, models.RoleEmployee)

	expiredToken := func() string {
		claims := jwt.MapClaims{
			"role": models.RoleModerator,
			"exp":  time.Now().Add(-time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(secretKey))
		return tokenString
	}()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   dto.Error
		expectedRole   string
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   dto.Error{Message: "Authorization header required"},
		},
		{
			name:           "invalid header format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   dto.Error{Message: "Authorization header must be 'Bearer <token>'"},
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   dto.Error{Message: "Invalid token"},
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + expiredToken,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   dto.Error{Message: "Invalid token"},
		},
		{
			name:           "valid moderator token",
			authHeader:     "Bearer " + validModeratorToken,
			expectedStatus: http.StatusOK,
			expectedRole:   models.RoleModerator,
		},
		{
			name:           "valid employee token",
			authHeader:     "Bearer " + validEmployeeToken,
			expectedStatus: http.StatusOK,
			expectedRole:   models.RoleEmployee,
		},
		{
			name: "token with invalid role type",
			authHeader: "Bearer " + func() string {
				claims := jwt.MapClaims{
					"role": 123,
					"exp":  time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secretKey))
				return tokenString
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   dto.Error{Message: "Role claim missing or invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := AuthMiddleware(cfg, &logger)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				ctx.Request.Header.Set("Authorization", tt.authHeader)
			}

			middleware(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus != http.StatusOK {
				var response dto.Error
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			} else {
				role, exists := ctx.Get("role")
				assert.True(t, exists)
				assert.Equal(t, tt.expectedRole, role)
			}
		})
	}
}

func TestRoleMiddlewares(t *testing.T) {
	tests := []struct {
		name           string
		middleware     gin.HandlerFunc
		setupContext   func(*gin.Context)
		expectedStatus int
	}{
		{
			name:       "ModeratorOnly - success",
			middleware: ModeratorOnly,
			setupContext: func(c *gin.Context) {
				c.Set("role", models.RoleModerator)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "ModeratorOnly - wrong role",
			middleware: ModeratorOnly,
			setupContext: func(c *gin.Context) {
				c.Set("role", models.RoleEmployee)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "ModeratorOnly - missing role",
			middleware: ModeratorOnly,
			setupContext: func(c *gin.Context) {

			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "EmployeeOnly - success",
			middleware: EmployeeOnly,
			setupContext: func(c *gin.Context) {
				c.Set("role", models.RoleEmployee)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "EmployeeOnly - wrong role",
			middleware: EmployeeOnly,
			setupContext: func(c *gin.Context) {
				c.Set("role", models.RoleModerator)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "EmployeeOnly - missing role",
			middleware: EmployeeOnly,
			setupContext: func(c *gin.Context) {

			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("GET", "/", nil)

			if tt.setupContext != nil {
				tt.setupContext(ctx)
			}

			tt.middleware(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus != http.StatusOK {
				var response dto.Error
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, dto.Error{Message: "Forbidden"}, response)
			}
		})
	}
}

var jwtParseWithClaims = jwt.ParseWithClaims

func init() {

	jwtParseWithClaims = jwt.ParseWithClaims
}
