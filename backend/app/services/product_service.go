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
	ErrProductInvalidID           = errors.New("product ID must be greater than zero")
	ErrProductCodeRequired        = errors.New("product code is required")
	ErrProductNameRequired        = errors.New("product name is required")
	ErrProductUnitRequired        = errors.New("product unit is required")
	ErrProductCategoryRequired    = errors.New("product category is required")
	ErrProductMinimumStockInvalid = errors.New("minimum stock must be zero or greater")
	ErrProductDuplicateCode       = errors.New("product code is already in use")
	ErrProductCategoryNotFound    = errors.New("product category not found")
	ErrProductInUse               = repositories.ErrProductInUse
)

type ProductInput struct {
	CategoryID   uint64 `json:"category_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Unit         string `json:"unit"`
	MinimumStock int64  `json:"minimum_stock"`
	IsActive     *bool  `json:"is_active"`
}
type ProductService struct{ db contractsorm.Orm }

func NewProductService(db contractsorm.Orm) *ProductService { return &ProductService{db: db} }
func (s *ProductService) Create(ctx context.Context, input ProductInput, actorID uint64) (*ProductResponse, error) {
	input, err := normalizeProduct(input)
	if err != nil {
		return nil, err
	}
	if actorID == 0 {
		return nil, ErrAuditLogInvalidUserID
	}
	var out *ProductResponse
	err = s.withContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewProductRepository(tx)
		if _, e := repositories.NewCategoryRepository(tx).FindByID(input.CategoryID); e != nil {
			if errors.Is(e, frameworkerrors.OrmRecordNotFound) {
				return ErrProductCategoryNotFound
			}
			return e
		}
		active := true
		if input.IsActive != nil {
			active = *input.IsActive
		}
		row := &models.Product{CategoryID: input.CategoryID, Code: input.Code, Name: input.Name, Description: input.Description, Unit: input.Unit, MinimumStock: input.MinimumStock, IsActive: active}
		if e := repo.Create(row); e != nil {
			if isDuplicateEntryError(e) {
				return ErrProductDuplicateCode
			}
			return e
		}
		if e := productAudit(tx, actorID, "CREATE_PRODUCT", row.ID, nil, productAuditValues(row)); e != nil {
			return e
		}
		mapped := productResponse(*row)
		out = &mapped
		return nil
	})
	return out, err
}
func (s *ProductService) Update(ctx context.Context, id uint64, input ProductInput, actorID uint64) (*ProductResponse, error) {
	if id == 0 {
		return nil, ErrProductInvalidID
	}
	input, err := normalizeProduct(input)
	if err != nil {
		return nil, err
	}
	if actorID == 0 {
		return nil, ErrAuditLogInvalidUserID
	}
	var out *ProductResponse
	err = s.withContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewProductRepository(tx)
		row, e := repo.FindByID(id)
		if e != nil {
			return e
		}
		if _, e = repositories.NewCategoryRepository(tx).FindByID(input.CategoryID); e != nil {
			if errors.Is(e, frameworkerrors.OrmRecordNotFound) {
				return ErrProductCategoryNotFound
			}
			return e
		}
		old := productAuditValues(row)
		active := row.IsActive
		if input.IsActive != nil {
			active = *input.IsActive
		}
		values := map[string]any{"category_id": input.CategoryID, "code": input.Code, "name": input.Name, "description": input.Description, "unit": input.Unit, "minimum_stock": input.MinimumStock, "is_active": active}
		if e = repo.Update(row, values); e != nil {
			if isDuplicateEntryError(e) {
				return ErrProductDuplicateCode
			}
			return e
		}
		if e = productAudit(tx, actorID, "UPDATE_PRODUCT", row.ID, old, values); e != nil {
			return e
		}
		mapped := ProductResponse{ID: uint64(row.ID), CategoryID: input.CategoryID, Code: input.Code, Name: input.Name, Description: input.Description, Unit: input.Unit, MinimumStock: input.MinimumStock, IsActive: active}
		out = &mapped
		return nil
	})
	return out, err
}
func (s *ProductService) Delete(ctx context.Context, id, actorID uint64) error {
	if id == 0 {
		return ErrProductInvalidID
	}
	if actorID == 0 {
		return ErrAuditLogInvalidUserID
	}
	return s.withContext(ctx).Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewProductRepository(tx)
		row, e := repo.FindByID(id)
		if e != nil {
			return e
		}
		if e = repo.HasReferences(id); e != nil {
			return e
		}
		old := productAuditValues(row)
		if e = repo.Delete(row); e != nil {
			return e
		}
		return productAudit(tx, actorID, "DELETE_PRODUCT", row.ID, old, map[string]any{"deleted": true})
	})
}
func (s *ProductService) withContext(ctx context.Context) contractsorm.Orm {
	if ctx != nil {
		return s.db.WithContext(ctx)
	}
	return s.db
}
func normalizeProduct(i ProductInput) (ProductInput, error) {
	i.Code = strings.TrimSpace(i.Code)
	i.Name = strings.TrimSpace(i.Name)
	i.Description = strings.TrimSpace(i.Description)
	i.Unit = strings.TrimSpace(i.Unit)
	if i.CategoryID == 0 {
		return i, ErrProductCategoryRequired
	}
	if i.Code == "" {
		return i, ErrProductCodeRequired
	}
	if i.Name == "" {
		return i, ErrProductNameRequired
	}
	if i.Unit == "" {
		return i, ErrProductUnitRequired
	}
	if i.MinimumStock < 0 {
		return i, ErrProductMinimumStockInvalid
	}
	return i, nil
}
func productAuditValues(r *models.Product) map[string]any {
	return map[string]any{"category_id": r.CategoryID, "code": r.Code, "name": r.Name, "description": r.Description, "unit": r.Unit, "minimum_stock": r.MinimumStock, "is_active": r.IsActive}
}
func productAudit(tx contractsorm.Query, actorID uint64, action string, id uint, old, new map[string]any) error {
	a, _ := json.Marshal(old)
	b, _ := json.Marshal(new)
	u := actorID
	return NewAuditLogService(repositories.NewAuditLogRepository(tx)).Create(&u, action, "Product", uint64(id), string(a), string(b), "", "")
}
