package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"

	"avito-pvz/internal/config"
	"avito-pvz/internal/dto"
	"avito-pvz/internal/models"
)

func AuthMiddleware(cfg *config.Config, logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn().Msg("Authorization header missing")
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Error{Message: "Authorization header required"})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn().Str("header", authHeader).Msg("Invalid Authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Error{Message: "Authorization header must be 'Bearer <token>'"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.SecretKey), nil
		})

		if err != nil || !token.Valid {
			logger.Warn().Err(err).Msg("Invalid token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Error{Message: "Invalid token"})
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			logger.Warn().Msg("Role claim missing or invalid")
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Error{Message: "Role claim missing or invalid"})
			return
		}

		c.Set("role", role)
		c.Next()
	}
}

func ModeratorOnly(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != models.RoleModerator {
		c.AbortWithStatusJSON(http.StatusForbidden, dto.Error{Message: "Forbidden"})
		return
	}
	c.Next()
}

func EmployeeOnly(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != models.RoleEmployee {
		c.AbortWithStatusJSON(http.StatusForbidden, dto.Error{Message: "Forbidden"})
		return
	}
	c.Next()
}
