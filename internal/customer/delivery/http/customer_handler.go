package http

import (
	"net/http"
	"odomosml/internal/customer/model"
	"odomosml/internal/customer/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CustomerHandler handles HTTP requests for customers
type CustomerHandler struct {
	service service.CustomerService
}

// NewCustomerHandler maakt een nieuwe CustomerHandler instantie
func NewCustomerHandler(service service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		service: service,
	}
}

// Helper functie om integer parameters te parsen
func parseIntParam(c *gin.Context, param string, defaultValue int) int {
	valueStr := c.Query(param)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil || value < 0 {
		return defaultValue
	}

	return value
}

// @Summary      Lijst van klanten ophalen
// @Description  Haalt een lijst van alle klanten op met optionele filters
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        page query int false "Paginanummer (default: 1)"
// @Param        pageSize query int false "Aantal items per pagina (default: 10, max: 100)"
// @Param        searchTerm query string false "Zoekterm voor naam of email"
// @Success      200  {object}  map[string]interface{} "Succesvol opgehaald"
// @Failure      400  {object}  map[string]string "Ongeldige parameters"
// @Failure      401  {object}  map[string]string "Niet geautoriseerd"
// @Failure      500  {object}  map[string]string "Server error"
// @Security     Bearer
// @Router       /klanten [get]
func (h *CustomerHandler) GetAll(c *gin.Context) {
	// Parse filter parameters
	page := parseIntParam(c, "page", 1)
	pageSize := parseIntParam(c, "page_size", 10)

	// Beperk pageSize om database overbelasting te voorkomen
	if pageSize > 100 {
		pageSize = 100
	}

	filter := model.CustomerFilter{
		SearchTerm: c.Query("zoekterm"),
		Page:       page,
		PageSize:   pageSize,
	}

	customers, total, err := h.service.GetAllCustomers(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Bereken paginering
	totalPages := (int(total) + filter.PageSize - 1) / filter.PageSize

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    customers,
		"pagination": gin.H{
			"current_page": filter.Page,
			"page_size":    filter.PageSize,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	})
}

// @Summary      Klant ophalen op ID
// @Description  Haalt een specifieke klant op basis van ID
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        id path string true "Klant ID"
// @Success      200  {object}  model.Customer "Succesvol opgehaald"
// @Failure      400  {object}  map[string]string "Ongeldig ID"
// @Failure      401  {object}  map[string]string "Niet geautoriseerd"
// @Failure      404  {object}  map[string]string "Klant niet gevonden"
// @Failure      500  {object}  map[string]string "Server error"
// @Security     Bearer
// @Router       /klanten/{id} [get]
func (h *CustomerHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	customer, err := h.service.GetCustomerByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    customer,
	})
}

// @Summary      Nieuwe klant aanmaken
// @Description  Maakt een nieuwe klant aan
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        customer body model.Customer true "Klant gegevens"
// @Success      201  {object}  model.Customer "Succesvol aangemaakt"
// @Failure      400  {object}  map[string]string "Ongeldige invoer"
// @Failure      401  {object}  map[string]string "Niet geautoriseerd"
// @Failure      409  {object}  map[string]string "Email bestaat al"
// @Failure      500  {object}  map[string]string "Server error"
// @Security     Bearer
// @Router       /klanten [post]
func (h *CustomerHandler) Create(c *gin.Context) {
	var customer model.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ongeldige request: " + err.Error(),
		})
		return
	}

	created, err := h.service.CreateCustomer(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    created,
	})
}

// @Summary      Klant bijwerken
// @Description  Werkt een bestaande klant bij
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        id path string true "Klant ID"
// @Param        customer body model.Customer true "Klant gegevens"
// @Success      200  {object}  model.Customer "Succesvol bijgewerkt"
// @Failure      400  {object}  map[string]string "Ongeldige invoer"
// @Failure      401  {object}  map[string]string "Niet geautoriseerd"
// @Failure      404  {object}  map[string]string "Klant niet gevonden"
// @Failure      409  {object}  map[string]string "Email bestaat al"
// @Failure      500  {object}  map[string]string "Server error"
// @Security     Bearer
// @Router       /klanten/{id} [put]
func (h *CustomerHandler) Update(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ongeldig ID",
		})
		return
	}

	var customer model.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ongeldige request: " + err.Error(),
		})
		return
	}

	customer.ID = uint(idInt)

	updated, err := h.service.UpdateCustomer(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updated,
	})
}

// @Summary      Klant gedeeltelijk bijwerken
// @Description  Werkt specifieke velden van een klant bij
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        id path string true "Klant ID"
// @Param        customer body object true "Klant velden om bij te werken"
// @Success      200  {object}  model.Customer "Succesvol bijgewerkt"
// @Failure      400  {object}  map[string]string "Ongeldige invoer"
// @Failure      401  {object}  map[string]string "Niet geautoriseerd"
// @Failure      404  {object}  map[string]string "Klant niet gevonden"
// @Failure      500  {object}  map[string]string "Server error"
// @Security     Bearer
// @Router       /klanten/{id} [patch]
func (h *CustomerHandler) PartialUpdate(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ongeldige request: " + err.Error(),
		})
		return
	}

	updated, err := h.service.PartialUpdateCustomer(id, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updated,
	})
}

// @Summary      Klant verwijderen
// @Description  Verwijdert een klant
// @Tags         customers
// @Accept       json
// @Produce      json
// @Param        id path string true "Klant ID"
// @Success      200  {object}  map[string]interface{} "Succesvol verwijderd"
// @Failure      400  {object}  map[string]string "Ongeldig ID"
// @Failure      401  {object}  map[string]string "Niet geautoriseerd"
// @Failure      404  {object}  map[string]string "Klant niet gevonden"
// @Failure      500  {object}  map[string]string "Server error"
// @Security     Bearer
// @Router       /klanten/{id} [delete]
func (h *CustomerHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	customerData, err := h.service.DeleteCustomer(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Sla customerData op in context voor audit logging
	c.Set("customerData", customerData)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Klant succesvol verwijderd",
	})
}
