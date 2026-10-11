package services

import (
	"context"
	"encoding/json"
	"errors"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	frameworkerrors "github.com/goravel/framework/errors"
	"goravel/app/models"
	"goravel/app/repositories"
	"strings"
)

var (
	ErrWarehouseInvalidID     = errors.New("warehouse ID must be greater than zero")
	ErrWarehouseCodeRequired  = errors.New("warehouse code is required")
	ErrWarehouseNameRequired  = errors.New("warehouse name is required")
	ErrWarehouseDuplicateCode = errors.New("warehouse code is already in use")
	ErrWarehouseInUse         = repositories.ErrWarehouseInUse
)

type WarehouseInput struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}
type WarehouseService struct{ db contractsorm.Orm }

func NewWarehouseService(db contractsorm.Orm) *WarehouseService { return &WarehouseService{db: db} }
func (s *WarehouseService) Create(ctx context.Context, i WarehouseInput, actor uint64) (*WarehouseResponse, error) {
	i, e := normalizeWarehouse(i)
	if e != nil {
		return nil, e
	}
	if actor == 0 {
		return nil, ErrAuditLogInvalidUserID
	}
	var out *WarehouseResponse
	e = s.db.WithContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewWarehouseRepository(tx)
		active := true
		if i.IsActive != nil {
			active = *i.IsActive
		}
		row := &models.Warehouse{Code: i.Code, Name: i.Name, Location: i.Location, Description: i.Description, IsActive: active}
		if e := repo.Create(row); e != nil {
			if isDuplicateEntryError(e) {
				return ErrWarehouseDuplicateCode
			}
			return e
		}
		if e := warehouseAudit(tx, actor, "CREATE_WAREHOUSE", row.ID, nil, warehouseValues(row)); e != nil {
			return e
		}
		v := warehouseResponse(*row)
		out = &v
		return nil
	})
	return out, e
}
func (s *WarehouseService) Update(ctx context.Context, id uint64, i WarehouseInput, actor uint64) (*WarehouseResponse, error) {
	if id == 0 {
		return nil, ErrWarehouseInvalidID
	}
	i, e := normalizeWarehouse(i)
	if e != nil {
		return nil, e
	}
	if actor == 0 {
		return nil, ErrAuditLogInvalidUserID
	}
	var out *WarehouseResponse
	e = s.db.WithContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewWarehouseRepository(tx)
		row, e := repo.FindByID(id)
		if e != nil {
			return e
		}
		old := warehouseValues(row)
		active := row.IsActive
		if i.IsActive != nil {
			active = *i.IsActive
		}
		v := map[string]any{"code": i.Code, "name": i.Name, "location": i.Location, "description": i.Description, "is_active": active}
		if e = repo.Update(row, v); e != nil {
			if isDuplicateEntryError(e) {
				return ErrWarehouseDuplicateCode
			}
			return e
		}
		if e = warehouseAudit(tx, actor, "UPDATE_WAREHOUSE", row.ID, old, v); e != nil {
			return e
		}
		out = &WarehouseResponse{ID: uint64(row.ID), Code: i.Code, Name: i.Name, Location: i.Location, Description: i.Description, IsActive: active}
		return nil
	})
	return out, e
}
func (s *WarehouseService) Delete(ctx context.Context, id, actor uint64) error {
	if id == 0 {
		return ErrWarehouseInvalidID
	}
	if actor == 0 {
		return ErrAuditLogInvalidUserID
	}
	return s.db.WithContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewWarehouseRepository(tx)
		row, e := repo.FindByID(id)
		if e != nil {
			return e
		}
		if e = repo.HasReferences(id); e != nil {
			return e
		}
		old := warehouseValues(row)
		if e = repo.Delete(row); e != nil {
			return e
		}
		return warehouseAudit(tx, actor, "DELETE_WAREHOUSE", row.ID, old, map[string]any{"deleted": true})
	})
}
func normalizeWarehouse(i WarehouseInput) (WarehouseInput, error) {
	i.Code = strings.TrimSpace(i.Code)
	i.Name = strings.TrimSpace(i.Name)
	i.Location = strings.TrimSpace(i.Location)
	i.Description = strings.TrimSpace(i.Description)
	if i.Code == "" {
		return i, ErrWarehouseCodeRequired
	}
	if i.Name == "" {
		return i, ErrWarehouseNameRequired
	}
	return i, nil
}
func warehouseValues(r *models.Warehouse) map[string]any {
	return map[string]any{"code": r.Code, "name": r.Name, "location": r.Location, "description": r.Description, "is_active": r.IsActive}
}
func warehouseAudit(tx contractsorm.Query, actor uint64, action string, id uint, old, new map[string]any) error {
	a, _ := json.Marshal(old)
	b, _ := json.Marshal(new)
	u := actor
	return NewAuditLogService(repositories.NewAuditLogRepository(tx)).Create(&u, action, "Warehouse", uint64(id), string(a), string(b), "", "")
}

var _ = frameworkerrors.OrmRecordNotFound
