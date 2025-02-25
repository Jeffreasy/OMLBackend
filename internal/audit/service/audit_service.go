package service

import (
	"odomosml/internal/audit/model"
	"odomosml/internal/audit/repository"
)

type AuditService interface {
	LogAction(userID uint, username string, actionType model.ActionType, entityType model.EntityType, entityID string, description string) error
	GetAuditLogs(filter model.AuditLogFilter) ([]model.AuditLog, int64, error)
}

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) LogAction(userID uint, username string, actionType model.ActionType, entityType model.EntityType, entityID string, description string) error {
	log := &model.AuditLog{
		UserID:      userID,
		Username:    username,
		ActionType:  actionType,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
	}
	return s.repo.Create(log)
}

func (s *auditService) GetAuditLogs(filter model.AuditLogFilter) ([]model.AuditLog, int64, error) {
	logs, total, err := s.repo.FindAll(filter)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
