package models

import "github.com/goravel/framework/database/orm"

type Category struct {
	orm.Model
	orm.SoftDeletes

	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	IsActive    bool   `gorm:"column:is_active"`

	Products []Product `gorm:"foreignKey:CategoryID;references:ID"`
}
