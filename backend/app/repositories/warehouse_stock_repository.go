package repositories

import (
	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// WarehouseStockRepository handles data access for warehouse stock records.
// Transaction boundaries are owned by the service layer and are represented by
// the query supplied when constructing the repository.
type WarehouseStockRepository struct {
	query contractsorm.Query
}

func NewWarehouseStockRepository(query contractsorm.Query) *WarehouseStockRepository {
	return &WarehouseStockRepository{query: query}
}

// FindByWarehouseAndProduct returns the stock row for a warehouse/product pair.
// It returns the ORM's not-found error when the row does not exist.
func (r *WarehouseStockRepository) FindByWarehouseAndProduct(warehouseID, productID uint64) (*models.WarehouseStock, error) {
	return r.findByWarehouseAndProduct(r.query, warehouseID, productID)
}

// FindByWarehouseAndProductForUpdate returns and locks the stock row for update.
// The supplied query must already be inside a database transaction.
func (r *WarehouseStockRepository) FindByWarehouseAndProductForUpdate(warehouseID, productID uint64) (*models.WarehouseStock, error) {
	return r.findByWarehouseAndProduct(r.query.LockForUpdate(), warehouseID, productID)
}

// FindByWarehouse returns all stock rows for a warehouse with their products
// eager loaded and ordered by product ID.
func (r *WarehouseStockRepository) FindByWarehouse(warehouseID uint64) ([]models.WarehouseStock, error) {
	var stocks []models.WarehouseStock

	if err := r.query.
		With("Product").
		Where("warehouse_id = ?", warehouseID).
		OrderBy("product_id", "asc").
		Find(&stocks); err != nil {
		return nil, err
	}

	return stocks, nil
}

func (r *WarehouseStockRepository) findByWarehouseAndProduct(query contractsorm.Query, warehouseID, productID uint64) (*models.WarehouseStock, error) {
	stock := &models.WarehouseStock{}

	if err := query.
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		FindOrFail(stock); err != nil {
		return nil, err
	}

	return stock, nil
}

func (r *WarehouseStockRepository) Create(stock *models.WarehouseStock) error {
	return r.query.Create(stock)
}

func (r *WarehouseStockRepository) UpdateQuantity(stock *models.WarehouseStock, quantity int64) error {
	_, err := r.query.Model(stock).Update("quantity", quantity)
	return err
}
