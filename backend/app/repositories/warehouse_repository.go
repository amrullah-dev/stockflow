package repositories

import (
	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// WarehouseRepository handles data access for warehouses.
type WarehouseRepository struct {
	query contractsorm.Query
}

func NewWarehouseRepository(query contractsorm.Query) *WarehouseRepository {
	return &WarehouseRepository{query: query}
}

func (r *WarehouseRepository) FindByID(warehouseID uint64) (*models.Warehouse, error) {
	warehouse := &models.Warehouse{}

	if err := r.query.
		Where("id = ?", warehouseID).
		FindOrFail(warehouse); err != nil {
		return nil, err
	}

	return warehouse, nil
}

func (r *WarehouseRepository) FindAll() ([]models.Warehouse, error) {
	var warehouses []models.Warehouse

	if err := r.query.
		OrderBy("id", "asc").
		Find(&warehouses); err != nil {
		return nil, err
	}

	return warehouses, nil
}
