package authHttp

import (
	"net/http"
	"odomosml/internal/auth/model"
	"odomosml/internal/auth/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq model.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.Login(loginReq.Email, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

// Register handles new user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var registerReq model.RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.Register(registerReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, token)
}

// Refresh generates a new token for a valid user
func (h *AuthHandler) Refresh(c *gin.Context) {
	// Get claims from context (set by auth middleware)
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "geen geldige authenticatie"})
		return
	}

	// Convert claims to proper type
	userClaims, ok := claims.(*model.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ongeldige claims data"})
		return
	}

	// Create new token with existing claims
	token, err := h.service.Login(userClaims.Email, "") // Special case for refresh
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: Implement logout logic
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
