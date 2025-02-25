package customerHttp

import (
	"net/http"
	"strconv"

	"odomosml/internal/customer/model"
	"odomosml/internal/customer/service"

	"github.com/gin-gonic/gin"
)

// CustomerHandler bevat de service voor klanten
type CustomerHandler struct {
	service service.CustomerService
}

// NewCustomerHandler creëert een nieuwe CustomerHandler met een meegegeven service
func NewCustomerHandler(svc service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		service: svc,
	}
}

// GetAll haalt alle klanten op met de meegegeven filteropties
func (h *CustomerHandler) GetAll(c *gin.Context) {
	filter := model.CustomerFilter{
		SearchTerm: c.Query("zoekterm"),
		Page:       parseIntParam(c.Query("page"), 1),
		PageSize:   parseIntParam(c.Query("pageSize"), 10),
		SortBy:     c.Query("sortBy"),
		SortOrder:  c.Query("sortOrder"),
	}

	customers, err := h.service.GetAllCustomers(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customers)
}

// GetByID haalt een enkele klant op aan de hand van het ID
func (h *CustomerHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	customer, err := h.service.GetCustomerByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customer)
}

// Create maakt een nieuwe klant aan
func (h *CustomerHandler) Create(c *gin.Context) {
	var customer model.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.service.CreateCustomer(customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// Update voert een volledige update uit van een klant
func (h *CustomerHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var customer model.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.UpdateCustomer(id, customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// PartialUpdate voert een gedeeltelijke update uit van een klant
func (h *CustomerHandler) PartialUpdate(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.PartialUpdateCustomer(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Delete verwijdert een klant
func (h *CustomerHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.DeleteCustomer(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Klant succesvol verwijderd"})
}

// parseIntParam parse een string naar een int, met een default-waarde als de parameter leeg is of niet geparseerd kan worden
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(param)
	if err != nil {
		return defaultValue
	}
	return value
}
