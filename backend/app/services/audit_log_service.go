package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"goravel/app/models"
	"goravel/app/repositories"
)

var (
	ErrAuditLogRepositoryNil    = errors.New("audit-log repository is nil")
	ErrAuditLogInvalidAction    = errors.New("audit-log action cannot be empty")
	ErrAuditLogInvalidSubject   = errors.New("audit-log subject type cannot be empty")
	ErrAuditLogInvalidSubjectID = errors.New("audit-log subject ID must be greater than zero")
	ErrAuditLogInvalidUserID    = errors.New("audit-log user ID must be greater than zero")
	ErrAuditLogInvalidDateRange = errors.New("audit-log date range is invalid")
)

type AuditLogService struct {
	repository *repositories.AuditLogRepository
}

func NewAuditLogService(repository *repositories.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repository: repository}
}

func (s *AuditLogService) Create(
	userID *uint64,
	action string,
	subjectType string,
	subjectID uint64,
	oldValues string,
	newValues string,
	ipAddress string,
	userAgent string,
) error {
	if err := s.validateRepository(); err != nil {
		return err
	}
	if err := validateAuditLogAction(action); err != nil {
		return err
	}
	if err := validateAuditLogSubject(subjectType, subjectID); err != nil {
		return err
	}

	if userID != nil && *userID == 0 {
		return ErrAuditLogInvalidUserID
	}

	log := &models.AuditLog{
		UserID:      userID,
		Action:      action,
		SubjectType: subjectType,
		SubjectID:   subjectID,
		OldValues:   oldValues,
		NewValues:   newValues,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}

	if err := s.repository.Create(log); err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}

	return nil
}

func (s *AuditLogService) FindBySubject(subjectType string, subjectID uint64) ([]models.AuditLog, error) {
	if err := s.validateRepository(); err != nil {
		return nil, err
	}
	if err := validateAuditLogSubject(subjectType, subjectID); err != nil {
		return nil, err
	}

	logs, err := s.repository.FindBySubject(subjectType, subjectID)
	if err != nil {
		return nil, fmt.Errorf("find audit logs by subject: %w", err)
	}

	return logs, nil
}

func (s *AuditLogService) FindByUser(userID uint64) ([]models.AuditLog, error) {
	if err := s.validateRepository(); err != nil {
		return nil, err
	}
	if userID == 0 {
		return nil, ErrAuditLogInvalidUserID
	}

	logs, err := s.repository.FindByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("find audit logs by user: %w", err)
	}

	return logs, nil
}

func (s *AuditLogService) FindByAction(action string) ([]models.AuditLog, error) {
	if err := s.validateRepository(); err != nil {
		return nil, err
	}
	if err := validateAuditLogAction(action); err != nil {
		return nil, err
	}

	logs, err := s.repository.FindByAction(action)
	if err != nil {
		return nil, fmt.Errorf("find audit logs by action: %w", err)
	}

	return logs, nil
}

func (s *AuditLogService) FindByDateRange(start, end time.Time) ([]models.AuditLog, error) {
	if err := s.validateRepository(); err != nil {
		return nil, err
	}
	if start.After(end) {
		return nil, ErrAuditLogInvalidDateRange
	}

	logs, err := s.repository.FindByDateRange(start, end)
	if err != nil {
		return nil, fmt.Errorf("find audit logs by date range: %w", err)
	}

	return logs, nil
}

func (s *AuditLogService) validateRepository() error {
	if s == nil || s.repository == nil {
		return ErrAuditLogRepositoryNil
	}

	return nil
}

func validateAuditLogAction(action string) error {
	if strings.TrimSpace(action) == "" {
		return ErrAuditLogInvalidAction
	}

	return nil
}

func validateAuditLogSubject(subjectType string, subjectID uint64) error {
	if strings.TrimSpace(subjectType) == "" {
		return ErrAuditLogInvalidSubject
	}
	if subjectID == 0 {
		return ErrAuditLogInvalidSubjectID
	}

	return nil
}
