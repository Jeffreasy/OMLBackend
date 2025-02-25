package middleware

import (
	"net/http"
	"strings"

	"odomosml/internal/auth/service"
	userModel "odomosml/internal/user/model"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "geen autorisatie header"})
			c.Abort()
			return
		}

		// Check if the header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ongeldige autorisatie header"})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Store claims in context for later use
		c.Set("claims", claims)
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// RoleMiddleware checks if the user has the required role
func RoleMiddleware(requiredRoles ...userModel.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "geen rol gevonden"})
			c.Abort()
			return
		}

		userRole := userModel.Role(role.(string))
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "onvoldoende rechten"})
			c.Abort()
			return
		}

		c.Next()
	}
}
