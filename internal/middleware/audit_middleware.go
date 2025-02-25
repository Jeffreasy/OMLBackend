package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"odomosml/internal/audit/model"
	"odomosml/internal/audit/service"
	"odomosml/internal/customer/repository"
	userRepo "odomosml/internal/user/repository"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ResponseWithID struct {
	ID uint `json:"id"`
}

type AuditMiddlewareConfig struct {
	AuditService service.AuditService
	CustomerRepo repository.CustomerRepository
	UserRepo     userRepo.UserRepository
}

// getEntityName vertaalt entity types naar Nederlandse namen
func getEntityName(entityType model.EntityType) string {
	switch entityType {
	case model.EntityCustomer:
		return "Klant"
	case model.EntityUser:
		return "Gebruiker"
	default:
		return string(entityType)
	}
}

// compareAndGetChanges vergelijkt oude en nieuwe waardes en geeft de verschillen terug
func compareAndGetChanges(old, new map[string]interface{}) []string {
	if old == nil || new == nil {
		return []string{}
	}

	var changes []string
	for key, oldValue := range old {
		if newValue, exists := new[key]; exists {
			// Skip empty values en password
			if key == "password" || oldValue == nil || newValue == nil {
				continue
			}
			// Converteer waardes naar strings voor vergelijking
			oldStr := fmt.Sprintf("%v", oldValue)
			newStr := fmt.Sprintf("%v", newValue)
			if oldStr != newStr {
				changes = append(changes, fmt.Sprintf("%s: '%v' → '%v'", key, oldValue, newValue))
			}
		}
	}

	if len(changes) == 0 {
		return []string{"geen wijzigingen"}
	}

	return changes
}

// AuditMiddleware is een middleware die acties logt voor audit doeleinden
type AuditMiddleware struct {
	service        service.AuditService
	logGetRequests bool // Optie om GET requests te loggen
}

// NewAuditMiddleware maakt een nieuwe audit middleware
func NewAuditMiddleware(service service.AuditService) gin.HandlerFunc {
	middleware := &AuditMiddleware{
		service:        service,
		logGetRequests: false, // Standaard GET requests niet loggen
	}

	return middleware.Handle
}

// NewAuditMiddlewareWithOptions maakt een nieuwe audit middleware met opties
func NewAuditMiddlewareWithOptions(service service.AuditService, logGetRequests bool) gin.HandlerFunc {
	middleware := &AuditMiddleware{
		service:        service,
		logGetRequests: logGetRequests,
	}

	return middleware.Handle
}

// Handle is de handler functie voor de middleware
func (m *AuditMiddleware) Handle(c *gin.Context) {
	// Skip audit logging voor bepaalde routes
	if strings.HasPrefix(c.Request.URL.Path, "/api/logs") {
		c.Next()
		return
	}

	// Skip GET requests tenzij expliciet ingeschakeld
	if c.Request.Method == "GET" && !m.logGetRequests {
		c.Next()
		return
	}

	// Haal gebruikersinformatie uit de context
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	// Bepaal actie type op basis van HTTP methode
	var actionType model.ActionType
	switch c.Request.Method {
	case "POST":
		actionType = model.ActionCreate
	case "PUT", "PATCH":
		actionType = model.ActionUpdate
	case "DELETE":
		actionType = model.ActionDelete
	case "GET":
		actionType = model.ActionRead
	default:
		actionType = model.ActionOther
	}

	// Bepaal entity type op basis van URL
	entityType := getEntityTypeFromPath(c.Request.URL.Path)

	// Haal entity ID uit URL als die er is
	entityID := getEntityIDFromPath(c.Request.URL.Path)

	// Maak een kopie van de context voor de response
	responseBodyWriter := &responseBodyWriter{
		ResponseWriter: c.Writer,
		body:           nil,
	}
	c.Writer = responseBodyWriter

	// Bewaar de oude data voor vergelijking (bij update/delete)
	var oldData map[string]interface{}
	var newData map[string]interface{}

	// Bij DELETE of UPDATE, haal de oude data op uit de context
	if actionType == model.ActionDelete {
		if userData, exists := c.Get("userData"); exists {
			if userDataMap, ok := userData.(map[string]interface{}); ok {
				oldData = userDataMap
			}
		}
	}

	// Voer de request uit
	c.Next()

	// Haal de nieuwe data op uit de request body (bij POST/PUT/PATCH)
	if actionType == model.ActionCreate || actionType == model.ActionUpdate {
		if c.Request.ContentLength > 0 {
			var requestBody map[string]interface{}
			if c.Request.Body != nil {
				if err := c.ShouldBindJSON(&requestBody); err == nil {
					newData = requestBody
				}
			}
		}
	}

	// Bouw de beschrijving op
	description := buildDescription(actionType, entityType, entityID, oldData, newData)

	// Maak een audit log entry
	auditLog := &model.AuditLog{
		UserID:      getUintValue(userID),
		Username:    getStringValue(username),
		ActionType:  actionType,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		OldData:     formatData(oldData),
		NewData:     formatData(newData),
		StatusCode:  responseBodyWriter.status,
		CreatedAt:   time.Now(),
	}

	// Log de audit entry
	if err := m.service.Create(auditLog); err != nil {
		log.Printf("Fout bij het loggen van audit: %v", err)
	}
}

// Helper functies

// responseBodyWriter is een wrapper rond gin.ResponseWriter die de response body opslaat
type responseBodyWriter struct {
	gin.ResponseWriter
	body   []byte
	status int
}

// Write implementeert de io.Writer interface
func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

// WriteHeader overschrijft de WriteHeader methode om de status code op te slaan
func (w *responseBodyWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// getEntityTypeFromPath haalt het entity type uit het pad
func getEntityTypeFromPath(path string) model.EntityType {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 {
		switch parts[2] {
		case "users":
			return model.EntityUser
		case "klanten":
			return model.EntityCustomer
		case "auth":
			return model.EntityAuth
		}
	}
	return model.EntityType("unknown")
}

// getEntityIDFromPath haalt het entity ID uit het pad
func getEntityIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		// Controleer of het laatste deel een ID is
		if _, err := strconv.Atoi(parts[3]); err == nil {
			return parts[3]
		}
	}
	return ""
}

// buildDescription bouwt een beschrijving op voor de audit log
func buildDescription(actionType model.ActionType, entityType model.EntityType, entityID string, oldData, newData map[string]interface{}) string {
	// Haal de Nederlandse naam van het entity type op
	entityName := getEntityName(entityType)

	switch actionType {
	case model.ActionCreate:
		switch entityType {
		case model.EntityUser:
			if newData != nil {
				return fmt.Sprintf("Nieuwe %s aangemaakt: %s (%s)",
					entityName,
					getStringFromMap(newData, "username"),
					getStringFromMap(newData, "email"))
			}
			return fmt.Sprintf("Nieuwe %s aangemaakt", entityName)
		case model.EntityCustomer:
			if newData != nil {
				return fmt.Sprintf("Nieuwe %s aangemaakt: %s",
					entityName,
					getStringFromMap(newData, "name"))
			}
			return fmt.Sprintf("Nieuwe %s aangemaakt", entityName)
		case model.EntityAuth:
			return "Token vernieuwd"
		}
	case model.ActionUpdate:
		switch entityType {
		case model.EntityUser:
			changes := compareAndGetChanges(oldData, newData)
			if len(changes) > 0 {
				return fmt.Sprintf("%s bijgewerkt (ID: %s): %s",
					entityName, entityID, strings.Join(changes, ", "))
			}
			return fmt.Sprintf("%s bijgewerkt (ID: %s)", entityName, entityID)
		case model.EntityCustomer:
			changes := compareAndGetChanges(oldData, newData)
			if len(changes) > 0 {
				return fmt.Sprintf("%s bijgewerkt (ID: %s): %s",
					entityName, entityID, strings.Join(changes, ", "))
			}
			return fmt.Sprintf("%s bijgewerkt (ID: %s)", entityName, entityID)
		}
	case model.ActionDelete:
		switch entityType {
		case model.EntityUser:
			if oldData != nil {
				return fmt.Sprintf("%s verwijderd (ID: %s) - gebruiker: %s (%s)",
					entityName,
					entityID,
					getStringFromMap(oldData, "username"),
					getStringFromMap(oldData, "email"))
			}
			return fmt.Sprintf("%s verwijderd (ID: %s)", entityName, entityID)
		case model.EntityCustomer:
			if oldData != nil {
				return fmt.Sprintf("%s verwijderd (ID: %s) - naam: %s",
					entityName,
					entityID,
					getStringFromMap(oldData, "name"))
			}
			return fmt.Sprintf("%s verwijderd (ID: %s)", entityName, entityID)
		}
	case model.ActionRead:
		switch entityType {
		case model.EntityUser:
			if entityID != "" {
				return fmt.Sprintf("%s opgevraagd (ID: %s)", entityName, entityID)
			}
			return fmt.Sprintf("%slijst opgevraagd", entityName)
		case model.EntityCustomer:
			if entityID != "" {
				return fmt.Sprintf("%s opgevraagd (ID: %s)", entityName, entityID)
			}
			return fmt.Sprintf("%slijst opgevraagd", entityName)
		}
	}

	return fmt.Sprintf("%s actie op %s %s", actionType, entityName, entityID)
}

// formatData formatteert data voor opslag in de audit log
func formatData(data map[string]interface{}) string {
	if data == nil {
		return ""
	}

	// Verwijder gevoelige data
	dataCopy := make(map[string]interface{})
	for k, v := range data {
		if k != "password" {
			dataCopy[k] = v
		}
	}

	// Converteer naar JSON
	jsonData, err := json.Marshal(dataCopy)
	if err != nil {
		return "{}"
	}

	return string(jsonData)
}

// Helper functies voor type conversie

func getUintValue(value interface{}) uint {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case uint:
		return v
	case int:
		return uint(v)
	case float64:
		return uint(v)
	case string:
		if id, err := strconv.Atoi(v); err == nil {
			return uint(id)
		}
	}

	return 0
}

func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getStringFromMap(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}

	if value, exists := data[key]; exists {
		return getStringValue(value)
	}

	return ""
}
