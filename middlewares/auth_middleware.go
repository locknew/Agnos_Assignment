package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"AgnosAssignments/services"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		claims, err := parseBearer(authService, authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("claims", claims)
		c.Set("hospital", claims.Hospital)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func OptionalAuth(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if claims, err := parseBearer(authService, authHeader); err == nil {
				c.Set("claims", claims)
				c.Set("hospital", claims.Hospital)
				c.Set("username", claims.Username)
			}
		}
		c.Next()
	}
}

func parseBearer(authService *services.AuthService, authHeader string) (*services.StaffClaims, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, fmt.Errorf("invalid authorization header")
	}
	return authService.ParseToken(parts[1])
}
