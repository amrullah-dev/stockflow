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
	ErrSupplierInvalidID     = errors.New("supplier ID must be greater than zero")
	ErrSupplierCodeRequired  = errors.New("supplier code is required")
	ErrSupplierNameRequired  = errors.New("supplier name is required")
	ErrSupplierDuplicateCode = errors.New("supplier code is already in use")
)

type SupplierInput struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	IsActive *bool  `json:"is_active"`
}
type SupplierResponse struct {
	ID       uint64 `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	IsActive bool   `json:"is_active"`
}
type SupplierService struct{ db contractsorm.Orm }

func NewSupplierService(db contractsorm.Orm) *SupplierService { return &SupplierService{db: db} }

func (s *SupplierService) FindAll() ([]SupplierResponse, error) {
	rows, err := repositories.NewSupplierRepository(s.db.Query()).FindAll()
	if err != nil {
		return nil, err
	}
	out := make([]SupplierResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, supplierResponse(row))
	}
	return out, nil
}
func (s *SupplierService) FindByID(id uint64) (*SupplierResponse, error) {
	if id == 0 {
		return nil, ErrSupplierInvalidID
	}
	row, err := repositories.NewSupplierRepository(s.db.Query()).FindByID(id)
	if err != nil {
		return nil, err
	}
	out := supplierResponse(*row)
	return &out, nil
}

func (s *SupplierService) Create(ctx context.Context, input SupplierInput, actorID uint64) (*SupplierResponse, error) {
	input, err := normalizeSupplier(input)
	if err != nil {
		return nil, err
	}
	if actorID == 0 {
		return nil, ErrAuditLogInvalidUserID
	}
	var response *SupplierResponse
	err = s.withContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewSupplierRepository(tx)
		active := true
		if input.IsActive != nil {
			active = *input.IsActive
		}
		row := &models.Supplier{Code: input.Code, Name: input.Name, Phone: input.Phone, Email: input.Email, Address: input.Address, IsActive: active}
		if err := repo.Create(row); err != nil {
			if isDuplicateEntryError(err) {
				return ErrSupplierDuplicateCode
			}
			return err
		}
		if err := supplierAudit(tx, actorID, "CREATE_SUPPLIER", row.ID, nil, supplierAuditValues(row)); err != nil {
			return err
		}
		out := supplierResponse(*row)
		response = &out
		return nil
	})
	return response, err
}

func (s *SupplierService) Update(ctx context.Context, id uint64, input SupplierInput, actorID uint64) (*SupplierResponse, error) {
	if id == 0 {
		return nil, ErrSupplierInvalidID
	}
	var err error
	input, err = normalizeSupplier(input)
	if err != nil {
		return nil, err
	}
	if actorID == 0 {
		return nil, ErrAuditLogInvalidUserID
	}
	var response *SupplierResponse
	err = s.withContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewSupplierRepository(tx)
		row, err := repo.FindByID(id)
		if err != nil {
			return err
		}
		old := supplierAuditValues(row)
		active := row.IsActive
		if input.IsActive != nil {
			active = *input.IsActive
		}
		values := map[string]any{"code": input.Code, "name": input.Name, "phone": input.Phone, "email": input.Email, "address": input.Address, "is_active": active}
		if err := repo.Update(row, values); err != nil {
			if isDuplicateEntryError(err) {
				return ErrSupplierDuplicateCode
			}
			return err
		}
		if err := supplierAudit(tx, actorID, "UPDATE_SUPPLIER", row.ID, old, values); err != nil {
			return err
		}
		out := SupplierResponse{ID: uint64(row.ID), Code: input.Code, Name: input.Name, Phone: input.Phone, Email: input.Email, Address: input.Address, IsActive: active}
		response = &out
		return nil
	})
	return response, err
}

func (s *SupplierService) Delete(ctx context.Context, id uint64, actorID uint64) error {
	if id == 0 {
		return ErrSupplierInvalidID
	}
	if actorID == 0 {
		return ErrAuditLogInvalidUserID
	}
	return s.withContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewSupplierRepository(tx)
		row, err := repo.FindByID(id)
		if err != nil {
			return err
		}
		old := supplierAuditValues(row)
		if err := repo.Delete(row); err != nil {
			return err
		}
		return supplierAudit(tx, actorID, "DELETE_SUPPLIER", row.ID, old, map[string]any{"deleted": true})
	})
}
func (s *SupplierService) withContext(ctx context.Context) contractsorm.Orm {
	if ctx != nil {
		return s.db.WithContext(ctx)
	}
	return s.db
}
func normalizeSupplier(input SupplierInput) (SupplierInput, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Email = strings.TrimSpace(input.Email)
	input.Address = strings.TrimSpace(input.Address)
	if input.Code == "" {
		return input, ErrSupplierCodeRequired
	}
	if input.Name == "" {
		return input, ErrSupplierNameRequired
	}
	return input, nil
}
func supplierResponse(row models.Supplier) SupplierResponse {
	return SupplierResponse{ID: uint64(row.ID), Code: row.Code, Name: row.Name, Phone: row.Phone, Email: row.Email, Address: row.Address, IsActive: row.IsActive}
}
func supplierAuditValues(row *models.Supplier) map[string]any {
	return map[string]any{"code": row.Code, "name": row.Name, "phone": row.Phone, "email": row.Email, "address": row.Address, "is_active": row.IsActive}
}
func supplierAudit(tx contractsorm.Query, actorID uint64, action string, subjectID uint, oldValues, newValues map[string]any) error {
	oldJSON, _ := json.Marshal(oldValues)
	newJSON, _ := json.Marshal(newValues)
	userID := actorID
	return NewAuditLogService(repositories.NewAuditLogRepository(tx)).Create(&userID, action, "Supplier", uint64(subjectID), string(oldJSON), string(newJSON), "", "")
}
