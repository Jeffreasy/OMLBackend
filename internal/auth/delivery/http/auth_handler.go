package http

import (
	"net/http"
	"odomosml/internal/auth/model"
	"odomosml/internal/auth/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	service service.AuthService
}

// NewAuthHandler maakt een nieuwe AuthHandler instantie
func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// @Summary      Inloggen
// @Description  Authenticeer een gebruiker en krijg een JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body LoginRequest true "Login gegevens"
// @Success      200  {object}  map[string]string "JWT token"
// @Failure      400  {object}  map[string]string "Ongeldige invoer"
// @Failure      401  {object}  map[string]string "Ongeldige inloggegevens"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq model.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ongeldige request: " + err.Error(),
		})
		return
	}

	token, err := h.service.Login(loginReq.Email, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    token,
	})
}

// Register handelt het registreren van nieuwe gebruikers
func (h *AuthHandler) Register(c *gin.Context) {
	var registerReq model.RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ongeldige request: " + err.Error(),
		})
		return
	}

	token, err := h.service.Register(registerReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    token,
	})
}

// @Summary      Token vernieuwen
// @Description  Vernieuw een JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string "Nieuw JWT token"
// @Failure      401  {object}  map[string]string "Ongeldige of verlopen token"
// @Failure      500  {object}  map[string]string "Server error"
// @Security     Bearer
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	// Haal claims uit context (gezet door AuthMiddleware)
	claimsInterface, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Geen geldige token gevonden",
		})
		return
	}

	userClaims, ok := claimsInterface.(*model.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Ongeldige token claims",
		})
		return
	}

	token, err := h.service.RefreshToken(userClaims)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    token,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: Implement logout logic
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// LoginRequest represents the login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}
