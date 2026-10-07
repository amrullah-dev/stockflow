package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"goravel/app/models"
	"goravel/app/repositories"
)

var (
	ErrCategoryInvalidID    = errors.New("category ID must be greater than zero")
	ErrCategoryNameRequired = errors.New("category name is required")
	ErrCategoryDuplicate    = errors.New("category name is already in use")
)

type CategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type CategoryResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type CategoryService struct{ db contractsorm.Orm }

func NewCategoryService(db contractsorm.Orm) *CategoryService { return &CategoryService{db: db} }

func (s *CategoryService) FindAll() ([]CategoryResponse, error) {
	categories, err := repositories.NewCategoryRepository(s.db.Query()).FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		result = append(result, categoryResponse(category))
	}
	return result, nil
}

func (s *CategoryService) FindByID(id uint64) (*CategoryResponse, error) {
	if id == 0 {
		return nil, ErrCategoryInvalidID
	}
	category, err := repositories.NewCategoryRepository(s.db.Query()).FindByID(id)
	if err != nil {
		return nil, err
	}
	response := categoryResponse(*category)
	return &response, nil
}

func (s *CategoryService) Create(ctx context.Context, input CategoryInput, actorID uint64) (*CategoryResponse, error) {
	name, err := validCategoryName(input.Name)
	if err != nil {
		return nil, err
	}
	if actorID == 0 {
		return nil, ErrAuditLogInvalidUserID
	}
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	var response *CategoryResponse
	err = s.withContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewCategoryRepository(tx)
		category := &models.Category{Name: name, Description: strings.TrimSpace(input.Description), IsActive: active}
		if err := repo.Create(category); err != nil {
			if isDuplicateEntryError(err) {
				return ErrCategoryDuplicate
			}
			return err
		}
		if err := categoryAudit(tx, actorID, "CREATE_CATEGORY", category.ID, nil, categoryAuditValues(category)); err != nil {
			return err
		}
		mapped := categoryResponse(*category)
		response = &mapped
		return nil
	})
	return response, err
}

func (s *CategoryService) Update(ctx context.Context, id uint64, input CategoryInput, actorID uint64) (*CategoryResponse, error) {
	if id == 0 {
		return nil, ErrCategoryInvalidID
	}
	name, err := validCategoryName(input.Name)
	if err != nil {
		return nil, err
	}
	if actorID == 0 {
		return nil, ErrAuditLogInvalidUserID
	}
	var response *CategoryResponse
	err = s.withContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewCategoryRepository(tx)
		category, err := repo.FindByID(id)
		if err != nil {
			return err
		}
		old := categoryAuditValues(category)
		active := category.IsActive
		if input.IsActive != nil {
			active = *input.IsActive
		}
		if err := repo.Update(category, map[string]any{"name": name, "description": strings.TrimSpace(input.Description), "is_active": active}); err != nil {
			if isDuplicateEntryError(err) {
				return ErrCategoryDuplicate
			}
			return err
		}
		if err := categoryAudit(tx, actorID, "UPDATE_CATEGORY", category.ID, old, map[string]any{"name": name, "description": strings.TrimSpace(input.Description), "is_active": active}); err != nil {
			return err
		}
		mapped := CategoryResponse{ID: uint64(category.ID), Name: name, Description: strings.TrimSpace(input.Description), IsActive: active}
		response = &mapped
		return nil
	})
	return response, err
}

func (s *CategoryService) Delete(ctx context.Context, id uint64, actorID uint64) error {
	if id == 0 {
		return ErrCategoryInvalidID
	}
	if actorID == 0 {
		return ErrAuditLogInvalidUserID
	}
	return s.withContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewCategoryRepository(tx)
		category, err := repo.FindByID(id)
		if err != nil {
			return err
		}
		old := categoryAuditValues(category)
		if err := repo.Delete(category); err != nil {
			return err
		}
		return categoryAudit(tx, actorID, "DELETE_CATEGORY", category.ID, old, map[string]any{"deleted": true})
	})
}

func (s *CategoryService) withContext(ctx context.Context) contractsorm.Orm {
	if ctx != nil {
		return s.db.WithContext(ctx)
	}
	return s.db
}
func validCategoryName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrCategoryNameRequired
	}
	return value, nil
}
func categoryResponse(category models.Category) CategoryResponse {
	return CategoryResponse{ID: uint64(category.ID), Name: category.Name, Description: category.Description, IsActive: category.IsActive}
}
func categoryAuditValues(category *models.Category) map[string]any {
	return map[string]any{"name": category.Name, "description": category.Description, "is_active": category.IsActive}
}
func categoryAudit(tx contractsorm.Query, actorID uint64, action string, subjectID uint, oldValues, newValues map[string]any) error {
	oldJSON, _ := json.Marshal(oldValues)
	newJSON, _ := json.Marshal(newValues)
	userID := actorID
	return NewAuditLogService(repositories.NewAuditLogRepository(tx)).Create(&userID, action, "Category", uint64(subjectID), string(oldJSON), string(newJSON), "", "")
}
