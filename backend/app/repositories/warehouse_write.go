package repositories

import (
	"errors"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"goravel/app/models"
)

var ErrWarehouseInUse = errors.New("warehouse is referenced by inventory or transaction history and cannot be deleted")

func (r *WarehouseRepository) Create(x *models.Warehouse) error { return r.query.Create(x) }
func (r *WarehouseRepository) Update(x *models.Warehouse, v map[string]any) error {
	_, e := r.query.Model(x).Update(v)
	return e
}
func (r *WarehouseRepository) Delete(x *models.Warehouse) error {
	_, e := r.query.Where("id = ?", x.ID).Delete(&models.Warehouse{})
	return e
}
func (r *WarehouseRepository) HasReferences(id uint64) error {
	for _, m := range []any{&models.WarehouseStock{}, &models.StockIn{}, &models.StockOut{}, &models.StockMovement{}} {
		n, e := r.query.Model(m).Where("warehouse_id = ?", id).Count()
		if e != nil {
			return e
		}
		if n > 0 {
			return ErrWarehouseInUse
		}
	}
	for _, m := range []any{&models.WarehouseTransfer{}} {
		n, e := r.query.Model(m).Where("from_warehouse_id = ? OR to_warehouse_id = ?", id, id).Count()
		if e != nil {
			return e
		}
		if n > 0 {
			return ErrWarehouseInUse
		}
	}
	return nil
}

var _ contractsorm.Query
