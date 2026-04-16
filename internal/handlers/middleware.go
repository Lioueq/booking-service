package middleware

import (
	"booking/internal/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type TokenParser interface {
	ParseToken(tokenStr string) (userID string, role string, err error)
}

func RequireAuth(parser TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.MakeError(
				response.ErrCodeUnauthorized,
				"unauthorized",
			))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.MakeError(
				response.ErrCodeUnauthorized,
				"unauthorized",
			))
			return
		}

		userID, role, err := parser.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.MakeError(
				response.ErrCodeUnauthorized,
				"unauthorized",
			))
			return
		}

		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.MakeError(
				response.ErrCodeUnauthorized,
				"unauthorized",
			))
			return
		}

		role, ok := v.(string)
		if !ok || role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.MakeError(
				response.ErrCodeUnauthorized,
				"unauthorized",
			))
			return
		}

		if !strings.EqualFold(role, requiredRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, response.MakeError(
				response.ErrCodeForbidden,
				"forbidden",
			))
			return
		}

		c.Next()
	}
}
