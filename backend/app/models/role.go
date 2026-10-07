package models

import "github.com/goravel/framework/database/orm"

type Role struct {
	orm.Model

	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`

	Users []User `gorm:"foreignKey:RoleID;references:ID"`
}
