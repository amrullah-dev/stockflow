package models

import "github.com/goravel/framework/database/orm"

type WarehouseTransferItem struct {
	orm.Model

	TransferID uint64 `gorm:"column:transfer_id"`
	ProductID  uint64 `gorm:"column:product_id"`
	Quantity   int64  `gorm:"column:quantity"`

	Transfer WarehouseTransfer `gorm:"foreignKey:TransferID;references:ID"`
	Product  Product           `gorm:"foreignKey:ProductID;references:ID"`
}
