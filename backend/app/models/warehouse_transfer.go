package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type WarehouseTransfer struct {
	orm.Model

	ReferenceNumber string     `gorm:"column:reference_number"`
	FromWarehouseID uint64     `gorm:"column:from_warehouse_id"`
	ToWarehouseID   uint64     `gorm:"column:to_warehouse_id"`
	TransactionDate time.Time  `gorm:"column:transaction_date"`
	Status          string     `gorm:"column:status"`
	Note            string     `gorm:"column:note"`
	CreatedBy       uint64     `gorm:"column:created_by"`
	PostedAt        *time.Time `gorm:"column:posted_at"`

	FromWarehouse Warehouse               `gorm:"foreignKey:FromWarehouseID;references:ID"`
	ToWarehouse   Warehouse               `gorm:"foreignKey:ToWarehouseID;references:ID"`
	CreatedByUser User                    `gorm:"foreignKey:CreatedBy;references:ID"`
	Items         []WarehouseTransferItem `gorm:"foreignKey:TransferID;references:ID"`
}
