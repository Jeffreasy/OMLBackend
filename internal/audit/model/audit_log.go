package model

import (
	"time"
)

// EntityType definieert het type entiteit dat wordt gelogd
type EntityType string

// ActionType definieert het type actie dat wordt gelogd
type ActionType string

// Entity types
const (
	EntityUser     EntityType = "user"
	EntityCustomer EntityType = "customer"
	EntityAuth     EntityType = "auth"
	EntityUnknown  EntityType = "unknown"
)

// Action types
const (
	ActionCreate ActionType = "create"
	ActionRead   ActionType = "read"
	ActionUpdate ActionType = "update"
	ActionDelete ActionType = "delete"
	ActionOther  ActionType = "other"
)

// AuditLog representeert een audit log entry
type AuditLog struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"index;not null"`
	Username    string     `json:"username" gorm:"size:100;index"`
	ActionType  ActionType `json:"action_type" gorm:"type:varchar(20);index;not null"`
	EntityType  EntityType `json:"entity_type" gorm:"size:50;index;not null"`
	EntityID    string     `json:"entity_id" gorm:"size:50;index"`
	Description string     `json:"description" gorm:"size:500"`
	OldData     string     `json:"old_data" gorm:"type:text"`
	NewData     string     `json:"new_data" gorm:"type:text"`
	StatusCode  int        `json:"status_code" gorm:"default:200"`
	CreatedAt   time.Time  `json:"created_at" gorm:"index;not null;default:CURRENT_TIMESTAMP"`
}

// AuditLogFilter definieert filters voor het ophalen van audit logs
type AuditLogFilter struct {
	UserID     uint       `json:"user_id" form:"user_id"`
	ActionType ActionType `json:"action_type" form:"action_type"`
	EntityType EntityType `json:"entity_type" form:"entity_type"`
	StartDate  time.Time  `json:"start_date" form:"start_date"`
	EndDate    time.Time  `json:"end_date" form:"end_date"`
	Page       int        `json:"page" form:"page"`
	PageSize   int        `json:"page_size" form:"page_size"`
}

// TableName specificeert de tabelnaam voor GORM
func (AuditLog) TableName() string {
	return "audit_logs"
}
