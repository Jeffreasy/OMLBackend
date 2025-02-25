package middleware

import (
	"net/http"
	"strings"

	"odomosml/internal/auth/service"
	userModel "odomosml/internal/user/model"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware controleert of de gebruiker geauthenticeerd is
// en zet de gebruikersinformatie in de context
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header is vereist",
			})
			c.Abort()
			return
		}

		// Controleer of de header begint met "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Ongeldig authorization header formaat",
			})
			c.Abort()
			return
		}

		// Haal de token uit de header
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Valideer de token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Ongeldige token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Zet gebruikersinformatie in de context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("userRole", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RoleMiddleware controleert of de gebruiker de vereiste rol heeft
func RoleMiddleware(requiredRoles ...userModel.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Haal de rol uit de context
		roleInterface, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Gebruiker niet geauthenticeerd",
			})
			c.Abort()
			return
		}

		// Converteer naar string en dan naar Role type
		userRole := userModel.Role(roleInterface.(string))

		// Controleer of de gebruiker een van de vereiste rollen heeft
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Onvoldoende rechten voor deze actie",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
