package http

import (
	"net/http"
	"odomosml/internal/user/model"
	"odomosml/internal/user/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	service service.UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// @Summary      Lijst van gebruikers ophalen
// @Description  Haalt een lijst van alle gebruikers op met optionele filters
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        page query int false "Paginanummer (default: 1)"
// @Param        pageSize query int false "Aantal items per pagina (default: 10, max: 100)"
// @Param        searchTerm query string false "Zoekterm voor gebruikersnaam of email"
// @Param        role query string false "Filter op rol (ADMIN/USER)"
// @Success      200  {object}  map[string]interface{} "{ data: []model.UserResponse, pagination: object }"
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Bearer
// @Router       /users [get]
func (h *UserHandler) GetAll(c *gin.Context) {
	filter := model.UserFilter{
		SearchTerm: c.Query("searchTerm"),
		Role:       model.Role(c.Query("role")),
		Page:       parseIntParam(c, "page", 1),
		PageSize:   parseIntParam(c, "pageSize", 10),
	}

	users, total, err := h.service.GetAllUsers(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert users to response objects
	userResponses := make([]model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data": userResponses,
		"pagination": gin.H{
			"total":    total,
			"page":     filter.Page,
			"pageSize": filter.PageSize,
			"lastPage": (total + int64(filter.PageSize) - 1) / int64(filter.PageSize),
		},
	})
}

// @Summary      Gebruiker ophalen op ID
// @Description  Haalt een specifieke gebruiker op basis van ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "Gebruiker ID"
// @Success      200  {object}  model.UserResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Bearer
// @Router       /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gebruiker niet gevonden"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// @Summary      Nieuwe gebruiker aanmaken
// @Description  Maakt een nieuwe gebruiker aan
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user body model.User true "Gebruiker gegevens"
// @Success      201  {object}  model.UserResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Bearer
// @Router       /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdUser, err := h.service.CreateUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdUser.ToResponse())
}

// @Summary      Gebruiker bijwerken
// @Description  Werkt een bestaande gebruiker bij
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "Gebruiker ID"
// @Param        user body model.User true "Gebruiker gegevens"
// @Success      200  {object}  model.UserResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Bearer
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse ID en converteer naar uint
	parsedID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ongeldig ID formaat"})
		return
	}
	user.ID = uint(parsedID)

	updatedUser, err := h.service.UpdateUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser.ToResponse())
}

// @Summary      Gebruiker verwijderen
// @Description  Verwijdert een gebruiker
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "Gebruiker ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     Bearer
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userData, err := h.service.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Sla de user data op in de context voor audit logging
	c.Set("userData", userData)

	c.JSON(http.StatusOK, gin.H{
		"message": "Gebruiker succesvol verwijderd",
		"data":    userData,
	})
}

// Helper function to parse integer parameters
func parseIntParam(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil || value < 1 {
		return defaultValue
	}
	return value
}
