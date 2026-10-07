package repositories

import (
	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// SupplierRepository handles data access for suppliers.
type SupplierRepository struct {
	query contractsorm.Query
}

func NewSupplierRepository(query contractsorm.Query) *SupplierRepository {
	return &SupplierRepository{query: query}
}

// FindByID returns a supplier by ID. Soft-deleted suppliers are automatically
// excluded by the ORM scope.
func (r *SupplierRepository) FindByID(id uint64) (*models.Supplier, error) {
	supplier := &models.Supplier{}

	if err := r.query.
		Where("id = ?", id).
		FindOrFail(supplier); err != nil {
		return nil, err
	}

	return supplier, nil
}

func (r *SupplierRepository) FindAll() ([]models.Supplier, error) {
	var suppliers []models.Supplier
	if err := r.query.OrderBy("id", "asc").Find(&suppliers); err != nil {
		return nil, err
	}
	return suppliers, nil
}

func (r *SupplierRepository) Create(supplier *models.Supplier) error {
	return r.query.Create(supplier)
}

func (r *SupplierRepository) Update(supplier *models.Supplier, values map[string]any) error {
	_, err := r.query.Model(supplier).Update(values)
	return err
}

func (r *SupplierRepository) Delete(supplier *models.Supplier) error {
	_, err := r.query.Where("id = ?", supplier.ID).Delete(&models.Supplier{})
	return err
}
