package models

import "github.com/goravel/framework/database/orm"

type Product struct {
	orm.Model
	orm.SoftDeletes

	CategoryID   uint64 `gorm:"column:category_id"`
	Code         string `gorm:"column:code"`
	Name         string `gorm:"column:name"`
	Description  string `gorm:"column:description"`
	Unit         string `gorm:"column:unit"`
	MinimumStock int64  `gorm:"column:minimum_stock"`
	IsActive     bool   `gorm:"column:is_active"`

	Category        Category                `gorm:"foreignKey:CategoryID;references:ID"`
	WarehouseStocks []WarehouseStock        `gorm:"foreignKey:ProductID;references:ID"`
	StockInItems    []StockInItem           `gorm:"foreignKey:ProductID;references:ID"`
	StockOutItems   []StockOutItem          `gorm:"foreignKey:ProductID;references:ID"`
	TransferItems   []WarehouseTransferItem `gorm:"foreignKey:ProductID;references:ID"`
	StockMovements  []StockMovement         `gorm:"foreignKey:ProductID;references:ID"`
}
