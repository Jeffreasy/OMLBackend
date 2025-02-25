package userHttp

import (
	"net/http"
	"odomosml/internal/user/model"
	"odomosml/internal/user/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// GetAll haalt alle gebruikers op
func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response objects
	var responses []model.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}
	c.JSON(http.StatusOK, responses)
}

// GetByID haalt een specifieke gebruiker op
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user.ToResponse())
}

// Create maakt een nieuwe gebruiker aan
func (h *UserHandler) Create(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.service.CreateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created.ToResponse())
}

// Update werkt een bestaande gebruiker bij
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdateUser(id, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated.ToResponse())
}

// Delete verwijdert een gebruiker
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	userData, err := h.service.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "gebruiker niet gevonden"})
		return
	}

	c.Set("userData", userData)
	c.JSON(http.StatusOK, gin.H{"message": "Gebruiker succesvol verwijderd"})
}
