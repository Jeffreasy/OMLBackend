package model

import "time"

type ActionType string

const (
	ActionCreate ActionType = "CREATE"
	ActionUpdate ActionType = "UPDATE"
	ActionDelete ActionType = "DELETE"
	ActionRead   ActionType = "READ"
)

type EntityType string

const (
	EntityCustomer EntityType = "CUSTOMER"
	EntityUser     EntityType = "USER"
	EntityAuth     EntityType = "AUTH"
)

type AuditLog struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id"`
	Username    string     `json:"username"`
	ActionType  ActionType `json:"action_type" gorm:"type:varchar(20)"`
	EntityType  EntityType `json:"entity_type" gorm:"type:varchar(20)"`
	EntityID    string     `json:"entity_id"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
}

type AuditLogFilter struct {
	UserID     uint
	ActionType ActionType
	EntityType EntityType
	StartDate  time.Time
	EndDate    time.Time
	Page       int
	PageSize   int
}
