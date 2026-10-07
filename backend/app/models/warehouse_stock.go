package models

import "github.com/goravel/framework/database/orm"

type WarehouseStock struct {
	orm.Model

	WarehouseID uint64 `gorm:"column:warehouse_id"`
	ProductID   uint64 `gorm:"column:product_id"`
	Quantity    int64  `gorm:"column:quantity"`

	Warehouse Warehouse `gorm:"foreignKey:WarehouseID;references:ID"`
	Product   Product   `gorm:"foreignKey:ProductID;references:ID"`
}
