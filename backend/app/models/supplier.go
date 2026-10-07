package models

import "github.com/goravel/framework/database/orm"

type Supplier struct {
	orm.Model
	orm.SoftDeletes

	Code     string `gorm:"column:code"`
	Name     string `gorm:"column:name"`
	Phone    string `gorm:"column:phone"`
	Email    string `gorm:"column:email"`
	Address  string `gorm:"column:address"`
	IsActive bool   `gorm:"column:is_active"`

	StockIns []StockIn `gorm:"foreignKey:SupplierID;references:ID"`
}
