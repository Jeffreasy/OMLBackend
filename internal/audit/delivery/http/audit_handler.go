package auditHttp

import (
	"net/http"
	"strconv"
	"time"

	"odomosml/internal/audit/model"
	"odomosml/internal/audit/service"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service service.AuditService
}

func NewAuditHandler(svc service.AuditService) *AuditHandler {
	return &AuditHandler{service: svc}
}

// GetLogs haalt audit logs op met filters
func (h *AuditHandler) GetLogs(c *gin.Context) {
	filter := model.AuditLogFilter{
		Page:     parseIntParam(c.Query("page"), 1),
		PageSize: parseIntParam(c.Query("pageSize"), 10),
	}

	// Parse filters
	if entityType := c.Query("entityType"); entityType != "" {
		if entityType == "klanten" {
			filter.EntityType = model.EntityCustomer
		} else if entityType == "users" {
			filter.EntityType = model.EntityUser
		}
	}

	if actionType := c.Query("actionType"); actionType != "" {
		switch actionType {
		case "create":
			filter.ActionType = model.ActionCreate
		case "update":
			filter.ActionType = model.ActionUpdate
		case "delete":
			filter.ActionType = model.ActionDelete
		}
	}

	// Parse datum filters
	if startDate := c.Query("startDate"); startDate != "" {
		if date, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = date
		}
	}

	if endDate := c.Query("endDate"); endDate != "" {
		if date, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = date
		}
	}

	// Haal logs op
	logs, total, err := h.service.GetAuditLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Bereken totaal aantal pagina's
	totalPages := (int(total) + filter.PageSize - 1) / filter.PageSize

	// Return met pagination info
	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"current_page": filter.Page,
			"page_size":    filter.PageSize,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	})
}

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
