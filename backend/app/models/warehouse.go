package models

import "github.com/goravel/framework/database/orm"

type Warehouse struct {
	orm.Model
	orm.SoftDeletes

	Code        string `gorm:"column:code"`
	Name        string `gorm:"column:name"`
	Location    string `gorm:"column:location"`
	Description string `gorm:"column:description"`
	IsActive    bool   `gorm:"column:is_active"`

	Stocks         []WarehouseStock    `gorm:"foreignKey:WarehouseID;references:ID"`
	StockIns       []StockIn           `gorm:"foreignKey:WarehouseID;references:ID"`
	StockOuts      []StockOut          `gorm:"foreignKey:WarehouseID;references:ID"`
	TransfersFrom  []WarehouseTransfer `gorm:"foreignKey:FromWarehouseID;references:ID"`
	TransfersTo    []WarehouseTransfer `gorm:"foreignKey:ToWarehouseID;references:ID"`
	StockMovements []StockMovement     `gorm:"foreignKey:WarehouseID;references:ID"`
}
