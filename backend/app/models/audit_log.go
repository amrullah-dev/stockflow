package models

import "github.com/goravel/framework/database/orm"

type AuditLog struct {
	orm.Model

	UserID      *uint64 `gorm:"column:user_id"`
	Action      string  `gorm:"column:action"`
	SubjectType string  `gorm:"column:subject_type"`
	SubjectID   uint64  `gorm:"column:subject_id"`
	OldValues   string  `gorm:"column:old_values"`
	NewValues   string  `gorm:"column:new_values"`
	IPAddress   string  `gorm:"column:ip_address"`
	UserAgent   string  `gorm:"column:user_agent"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
}
