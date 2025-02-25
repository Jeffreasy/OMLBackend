package middleware

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"odomosml/internal/audit/model"
	"odomosml/internal/audit/service"
	"odomosml/internal/customer/repository"
	userRepo "odomosml/internal/user/repository"
	"strings"

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
func compareAndGetChanges(old, new map[string]interface{}) string {
	if old == nil || new == nil {
		return ""
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
		return "geen wijzigingen"
	}

	return strings.Join(changes, ", ")
}

func NewAuditMiddleware(config AuditMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit voor GET requests en audit logs zelf
		if c.Request.Method == "GET" || strings.HasPrefix(c.Request.URL.Path, "/api/logs") {
			c.Next()
			return
		}

		// Speciale behandeling voor token refresh
		if c.Request.URL.Path == "/api/auth/refresh" && c.Request.Method == "POST" {
			// Haal user info uit de context
			userID, exists := c.Get("userID")
			if !exists {
				c.Next()
				return
			}

			username, exists := c.Get("username")
			if !exists {
				username = "unknown"
			}

			email, exists := c.Get("email")
			if !exists {
				email = "unknown"
			}

			description := fmt.Sprintf("Token vernieuwd voor gebruiker %s (%s)", username.(string), email.(string))

			// Log de refresh actie
			_ = config.AuditService.LogAction(
				userID.(uint),
				username.(string),
				model.ActionCreate,
				model.EntityAuth,
				fmt.Sprintf("USER_%d", userID.(uint)),
				description,
			)

			c.Next()
			return
		}

		// Lees en bewaar het originele request body voor updates
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = ioutil.ReadAll(c.Request.Body)
			// Reset het body voor de volgende handlers
			c.Request.Body = ioutil.NopCloser(strings.NewReader(string(requestBody)))
		}

		// Sla de response writer op
		blw := &bodyLogWriter{body: []byte{}, ResponseWriter: c.Writer}
		c.Writer = blw

		// Haal oude data op voor updates
		var oldData map[string]interface{}
		if (c.Request.Method == "PUT" || c.Request.Method == "PATCH") && c.Param("id") != "" {
			id := c.Param("id")
			var err error

			switch {
			case strings.Contains(c.Request.URL.Path, "/users"):
				oldData, err = config.UserRepo.GetUserForAudit(id)
			case strings.Contains(c.Request.URL.Path, "/klanten"):
				oldData, err = config.CustomerRepo.GetCustomerForAudit(id)
			}

			if err != nil {
				fmt.Printf("Error getting old data for audit: %v\n", err)
			}
		}

		// Voer de request uit
		c.Next()

		// Haal user info uit de context
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")

		// Als er geen user info is, gebruik system als fallback
		if userID == nil {
			userID = uint(0)
		}
		if username == nil {
			username = "system"
		}

		// Bepaal entity type
		var entityType model.EntityType
		if strings.Contains(c.Request.URL.Path, "/users") {
			entityType = model.EntityUser
		} else if strings.Contains(c.Request.URL.Path, "/klanten") {
			entityType = model.EntityCustomer
		}

		// Bepaal action type
		var actionType model.ActionType
		switch c.Request.Method {
		case "POST":
			actionType = model.ActionCreate
		case "PUT", "PATCH":
			actionType = model.ActionUpdate
		case "DELETE":
			actionType = model.ActionDelete
		}

		// Haal entity ID uit de URL of response
		entityID := c.Param("id")
		if entityID == "" && len(blw.body) > 0 {
			var resp ResponseWithID
			if err := json.Unmarshal(blw.body, &resp); err == nil && resp.ID > 0 {
				entityID = fmt.Sprintf("%d", resp.ID)
			}
		}

		// Parse request/response body voor meer details
		var newData map[string]interface{}
		if len(requestBody) > 0 {
			json.Unmarshal(requestBody, &newData)
		}

		// Maak een beschrijvende message
		entityName := getEntityName(entityType)
		var description string

		switch actionType {
		case model.ActionCreate:
			description = fmt.Sprintf("Nieuwe %s aangemaakt (ID: %s)", entityName, entityID)
			if newData != nil {
				if name, ok := newData["name"].(string); ok {
					description += fmt.Sprintf(" - %s", name)
				} else if username, ok := newData["username"].(string); ok {
					description += fmt.Sprintf(" - %s", username)
				}
			}
		case model.ActionUpdate:
			description = fmt.Sprintf("%s bijgewerkt (ID: %s)", entityName, entityID)
			if oldData != nil && newData != nil {
				changes := compareAndGetChanges(oldData, newData)
				if changes != "" {
					description += fmt.Sprintf(" - Wijzigingen: %s", changes)
				}
			}
		case model.ActionDelete:
			description = fmt.Sprintf("%s verwijderd (ID: %s)", entityName, entityID)
			if oldData != nil {
				if name, ok := oldData["name"].(string); ok {
					description += fmt.Sprintf(" - %s", name)
				} else if username, ok := oldData["username"].(string); ok {
					description += fmt.Sprintf(" - gebruiker: %s (%s)", username, oldData["email"].(string))
				}
			}
		}

		// Get the response body
		if strings.Contains(c.Request.URL.Path, "/api/users/") && c.Request.Method == "DELETE" {
			userData, exists := c.Get("userData")
			if exists {
				userDataMap, ok := userData.(map[string]interface{})
				if ok {
					username := userDataMap["username"].(string)
					email := userDataMap["email"].(string)
					description = fmt.Sprintf("Gebruiker verwijderd (ID: %s) - gebruiker: %s (%s)", entityID, username, email)
				}
			}
		}

		// Log de actie
		_ = config.AuditService.LogAction(
			userID.(uint),
			username.(string),
			actionType,
			entityType,
			entityID,
			description,
		)
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body = b
	return w.ResponseWriter.Write(b)
}
