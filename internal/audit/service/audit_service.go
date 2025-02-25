package service

import (
	"odomosml/internal/audit/model"
	"odomosml/internal/audit/repository"
)

// AuditService interface definieert de methodes voor audit logging
type AuditService interface {
	GetAuditLogs(filter model.AuditLogFilter) ([]model.AuditLog, int64, error)
	Create(log *model.AuditLog) error
}

// auditService implementeert de AuditService interface
type auditService struct {
	repo repository.AuditRepository
}

// NewAuditService maakt een nieuwe AuditService instantie
func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{
		repo: repo,
	}
}

// GetAuditLogs haalt audit logs op met filters
func (s *auditService) GetAuditLogs(filter model.AuditLogFilter) ([]model.AuditLog, int64, error) {
	// Valideer paginering
	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10 // Default page size
	}

	return s.repo.FindAll(filter)
}

// Create maakt een nieuwe audit log entry aan
func (s *auditService) Create(log *model.AuditLog) error {
	// Validatie
	if log.UserID == 0 {
		log.UserID = 1 // Default naar system user als geen gebruiker is opgegeven
	}

	return s.repo.Create(log)
}
