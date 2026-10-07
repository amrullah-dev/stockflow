package repositories

import (
	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// StockMovementRepository handles data access for stock movement records.
// Transaction boundaries are owned by the service layer and are represented by
// the query supplied when constructing the repository.
type StockMovementRepository struct {
	query contractsorm.Query
}

func NewStockMovementRepository(query contractsorm.Query) *StockMovementRepository {
	return &StockMovementRepository{query: query}
}

func (r *StockMovementRepository) Create(movement *models.StockMovement) error {
	return r.query.Create(movement)
}

func (r *StockMovementRepository) FindByWarehouseAndProduct(warehouseID, productID uint64) ([]models.StockMovement, error) {
	var movements []models.StockMovement

	err := r.query.
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		Find(&movements)
	if err != nil {
		return nil, err
	}

	return movements, nil
}

func (r *StockMovementRepository) FindByReference(referenceType string, referenceID uint64) ([]models.StockMovement, error) {
	var movements []models.StockMovement

	err := r.query.
		Where("reference_type = ? AND reference_id = ?", referenceType, referenceID).
		Find(&movements)
	if err != nil {
		return nil, err
	}

	return movements, nil
}
