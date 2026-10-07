package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type StockIn struct {
	orm.Model

	ReferenceNumber string     `gorm:"column:reference_number"`
	SupplierID      uint64     `gorm:"column:supplier_id"`
	WarehouseID     uint64     `gorm:"column:warehouse_id"`
	TransactionDate time.Time  `gorm:"column:transaction_date"`
	Status          string     `gorm:"column:status"`
	Note            string     `gorm:"column:note"`
	CreatedBy       uint64     `gorm:"column:created_by"`
	PostedAt        *time.Time `gorm:"column:posted_at"`

	Supplier      Supplier      `gorm:"foreignKey:SupplierID;references:ID"`
	Warehouse     Warehouse     `gorm:"foreignKey:WarehouseID;references:ID"`
	CreatedByUser User          `gorm:"foreignKey:CreatedBy;references:ID"`
	Items         []StockInItem `gorm:"foreignKey:StockInID;references:ID"`
}
