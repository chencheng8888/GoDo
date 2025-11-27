package model

import "time"

type User struct {
	UserName  string    `gorm:"column:user_name;type:varchar(255);primaryKey"`
	Password  string    `gorm:"column:password;type:varchar(255);not null"`
	UseShell  bool      `gorm:"column:use_shell;not null;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (u *User) TableName() string {
	return "users"
}
