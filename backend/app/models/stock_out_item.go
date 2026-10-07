package models

import "github.com/goravel/framework/database/orm"

type StockOutItem struct {
	orm.Model

	StockOutID uint64 `gorm:"column:stock_out_id"`
	ProductID  uint64 `gorm:"column:product_id"`
	Quantity   int64  `gorm:"column:quantity"`

	StockOut StockOut `gorm:"foreignKey:StockOutID;references:ID"`
	Product  Product  `gorm:"foreignKey:ProductID;references:ID"`
}
