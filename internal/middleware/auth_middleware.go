package middleware

import (
	"net/http"
	"strings"

	"odomosml/internal/auth/service"
	userModel "odomosml/internal/user/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("userRole", claims.Role)
		c.Set("claims", claims)

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

// NewJWTAuthMiddleware maakt een nieuwe JWT authenticatie middleware
func NewJWTAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check Bearer schema
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Parse en valideer token
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Haal claims op
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Sla user info op in context
		c.Set("userID", uint(claims["user_id"].(float64)))
		c.Set("username", claims["username"].(string))
		c.Set("email", claims["email"].(string))
		c.Set("role", claims["role"].(string))

		c.Next()
	}
}

// NewRoleMiddleware maakt een nieuwe role-based middleware
func NewRoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Gebruiker niet geauthenticeerd"})
			c.Abort()
			return
		}

		if role.(string) != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Onvoldoende rechten"})
			c.Abort()
			return
		}

		c.Next()
	}
}
