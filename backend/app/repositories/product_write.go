package repositories

import (
	"errors"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"goravel/app/models"
)

var ErrProductInUse = errors.New("product is referenced by inventory history and cannot be deleted")

func (r *ProductRepository) Create(row *models.Product) error { return r.query.Create(row) }
func (r *ProductRepository) Update(row *models.Product, values map[string]any) error {
	_, err := r.query.Model(row).Update(values)
	return err
}
func (r *ProductRepository) Delete(row *models.Product) error {
	_, err := r.query.Where("id = ?", row.ID).Delete(&models.Product{})
	return err
}
func (r *ProductRepository) HasReferences(id uint64) error {
	for _, model := range []any{&models.WarehouseStock{}, &models.StockInItem{}, &models.StockOutItem{}, &models.WarehouseTransferItem{}, &models.StockMovement{}} {
		count, err := r.query.Model(model).Where("product_id = ?", id).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrProductInUse
		}
	}
	return nil
}

var _ contractsorm.Query
