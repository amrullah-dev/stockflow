package models

import "github.com/goravel/framework/database/orm"

type StockInItem struct {
	orm.Model

	StockInID uint64  `gorm:"column:stock_in_id"`
	ProductID uint64  `gorm:"column:product_id"`
	Quantity  int64   `gorm:"column:quantity"`
	UnitPrice float64 `gorm:"column:unit_price"`
	Subtotal  float64 `gorm:"column:subtotal"`

	StockIn StockIn `gorm:"foreignKey:StockInID;references:ID"`
	Product Product `gorm:"foreignKey:ProductID;references:ID"`
}
