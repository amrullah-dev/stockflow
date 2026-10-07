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
