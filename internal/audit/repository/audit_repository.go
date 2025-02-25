package repository

import (
	"odomosml/internal/audit/model"

	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(log *model.AuditLog) error
	FindAll(filter model.AuditLogFilter) ([]model.AuditLog, int64, error)
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditRepository) FindAll(filter model.AuditLogFilter) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	query := r.db.Model(&model.AuditLog{})

	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.ActionType != "" {
		query = query.Where("action_type = ?", filter.ActionType)
	}

	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}

	if !filter.StartDate.IsZero() {
		query = query.Where("created_at >= ?", filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Add default sorting
	query = query.Order("created_at DESC")

	// Execute the query
	err = query.Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
