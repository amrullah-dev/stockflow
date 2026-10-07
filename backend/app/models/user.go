package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type User struct {
	orm.Model
	orm.SoftDeletes

	RoleID      uint64     `gorm:"column:role_id"`
	Name        string     `gorm:"column:name"`
	Email       string     `gorm:"column:email"`
	Password    string     `gorm:"column:password"`
	IsActive    bool       `gorm:"column:is_active"`
	LastLoginAt *time.Time `gorm:"column:last_login_at"`

	Role               Role                `gorm:"foreignKey:RoleID;references:ID"`
	StockIns           []StockIn           `gorm:"foreignKey:CreatedBy;references:ID"`
	StockOuts          []StockOut          `gorm:"foreignKey:CreatedBy;references:ID"`
	WarehouseTransfers []WarehouseTransfer `gorm:"foreignKey:CreatedBy;references:ID"`
	StockMovements     []StockMovement     `gorm:"foreignKey:CreatedBy;references:ID"`
	AuditLogs          []AuditLog          `gorm:"foreignKey:UserID;references:ID"`
}
