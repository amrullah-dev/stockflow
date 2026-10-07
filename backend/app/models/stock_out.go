package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type StockOut struct {
	orm.Model

	ReferenceNumber string     `gorm:"column:reference_number"`
	WarehouseID     uint64     `gorm:"column:warehouse_id"`
	TransactionDate time.Time  `gorm:"column:transaction_date"`
	Status          string     `gorm:"column:status"`
	Destination     string     `gorm:"column:destination"`
	Note            string     `gorm:"column:note"`
	CreatedBy       uint64     `gorm:"column:created_by"`
	PostedAt        *time.Time `gorm:"column:posted_at"`

	Warehouse     Warehouse      `gorm:"foreignKey:WarehouseID;references:ID"`
	CreatedByUser User           `gorm:"foreignKey:CreatedBy;references:ID"`
	Items         []StockOutItem `gorm:"foreignKey:StockOutID;references:ID"`
}
