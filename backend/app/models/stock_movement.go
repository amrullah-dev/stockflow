package models

import "github.com/goravel/framework/database/orm"

type StockMovement struct {
	orm.Model

	WarehouseID   uint64  `gorm:"column:warehouse_id"`
	ProductID     uint64  `gorm:"column:product_id"`
	Type          string  `gorm:"column:type"`
	Quantity      int64   `gorm:"column:quantity"`
	BalanceBefore int64   `gorm:"column:balance_before"`
	BalanceAfter  int64   `gorm:"column:balance_after"`
	ReferenceType string  `gorm:"column:reference_type"`
	ReferenceID   *uint64 `gorm:"column:reference_id"`
	Note          string  `gorm:"column:note"`
	CreatedBy     *uint64 `gorm:"column:created_by"`

	Warehouse     Warehouse `gorm:"foreignKey:WarehouseID;references:ID"`
	Product       Product   `gorm:"foreignKey:ProductID;references:ID"`
	CreatedByUser *User     `gorm:"foreignKey:CreatedBy;references:ID"`
}
