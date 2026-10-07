package services

import (
	"errors"

	"github.com/goravel/framework/support/carbon"

	"goravel/app/models"
	"goravel/app/repositories"
)

var ErrAuditLogQueryInvalidID = errors.New("audit log ID must be greater than zero")

type AuditLogUserResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AuditLogListResponse struct {
	ID          uint64                `json:"id"`
	Action      string                `json:"action"`
	SubjectType string                `json:"subject_type"`
	SubjectID   uint64                `json:"subject_id"`
	UserID      *uint64               `json:"user_id"`
	User        *AuditLogUserResponse `json:"user"`
	CreatedAt   *carbon.DateTime      `json:"created_at"`
}

type AuditLogDetailResponse struct {
	ID          uint64                `json:"id"`
	Action      string                `json:"action"`
	SubjectType string                `json:"subject_type"`
	SubjectID   uint64                `json:"subject_id"`
	OldValues   string                `json:"old_values"`
	NewValues   string                `json:"new_values"`
	IPAddress   string                `json:"ip_address"`
	UserAgent   string                `json:"user_agent"`
	UserID      *uint64               `json:"user_id"`
	User        *AuditLogUserResponse `json:"user"`
	CreatedAt   *carbon.DateTime      `json:"created_at"`
	UpdatedAt   *carbon.DateTime      `json:"updated_at"`
}

type AuditLogQueryService struct {
	repo *repositories.AuditLogRepository
}

func NewAuditLogQueryService(repo *repositories.AuditLogRepository) *AuditLogQueryService {
	return &AuditLogQueryService{repo: repo}
}

func (s *AuditLogQueryService) FindAll() ([]AuditLogListResponse, error) {
	logs, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]AuditLogListResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, auditLogListResponse(log))
	}

	return responses, nil
}

func (s *AuditLogQueryService) FindByID(id uint64) (*AuditLogDetailResponse, error) {
	if id == 0 {
		return nil, ErrAuditLogQueryInvalidID
	}

	log, err := s.repo.FindByIDForRead(id)
	if err != nil {
		return nil, err
	}

	response := auditLogDetailResponse(*log)
	return &response, nil
}

func auditLogUserResponse(user *models.User) *AuditLogUserResponse {
	if user == nil {
		return nil
	}

	return &AuditLogUserResponse{
		ID:    uint64(user.ID),
		Name:  user.Name,
		Email: user.Email,
	}
}

func auditLogListResponse(log models.AuditLog) AuditLogListResponse {
	return AuditLogListResponse{
		ID:          uint64(log.ID),
		Action:      log.Action,
		SubjectType: log.SubjectType,
		SubjectID:   log.SubjectID,
		UserID:      log.UserID,
		User:        auditLogUserResponse(log.User),
		CreatedAt:   log.CreatedAt,
	}
}

func auditLogDetailResponse(log models.AuditLog) AuditLogDetailResponse {
	return AuditLogDetailResponse{
		ID:          uint64(log.ID),
		Action:      log.Action,
		SubjectType: log.SubjectType,
		SubjectID:   log.SubjectID,
		OldValues:   log.OldValues,
		NewValues:   log.NewValues,
		IPAddress:   log.IPAddress,
		UserAgent:   log.UserAgent,
		UserID:      log.UserID,
		User:        auditLogUserResponse(log.User),
		CreatedAt:   log.CreatedAt,
		UpdatedAt:   log.UpdatedAt,
	}
}
