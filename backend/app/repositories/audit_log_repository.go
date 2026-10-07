package repositories

import (
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// AuditLogRepository handles data access for audit log records. Transaction
// boundaries are owned by the service layer and are represented by the query
// supplied when constructing the repository.
type AuditLogRepository struct {
	query contractsorm.Query
}

func NewAuditLogRepository(query contractsorm.Query) *AuditLogRepository {
	return &AuditLogRepository{query: query}
}

func (r *AuditLogRepository) Create(log *models.AuditLog) error {
	return r.query.Create(log)
}

func (r *AuditLogRepository) FindAll() ([]models.AuditLog, error) {
	var logs []models.AuditLog

	if err := r.query.
		With("User").
		OrderBy("created_at", "desc").
		OrderBy("id", "desc").
		Find(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *AuditLogRepository) FindByIDForRead(id uint64) (*models.AuditLog, error) {
	log := &models.AuditLog{}

	if err := r.query.
		With("User").
		Where("id = ?", id).
		FindOrFail(log); err != nil {
		return nil, err
	}

	return log, nil
}

func (r *AuditLogRepository) FindBySubject(subjectType string, subjectID uint64) ([]models.AuditLog, error) {
	var logs []models.AuditLog

	if err := r.query.
		Where("subject_type = ? AND subject_id = ?", subjectType, subjectID).
		Find(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *AuditLogRepository) FindByUser(userID uint64) ([]models.AuditLog, error) {
	var logs []models.AuditLog

	if err := r.query.
		Where("user_id = ?", userID).
		Find(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *AuditLogRepository) FindByAction(action string) ([]models.AuditLog, error) {
	var logs []models.AuditLog

	if err := r.query.
		Where("action = ?", action).
		Find(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *AuditLogRepository) FindByDateRange(start, end time.Time) ([]models.AuditLog, error) {
	var logs []models.AuditLog

	if err := r.query.
		Where("created_at BETWEEN ? AND ?", start, end).
		Find(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}
