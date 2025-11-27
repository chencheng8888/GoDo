package model

import "time"

type User struct {
	UserName  string    `gorm:"column:user_name;type:varchar(255);primaryKey"`
	Password  string    `gorm:"column:password;type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null"`
}

func (u *User) TableName() string {
	return "users"
}
