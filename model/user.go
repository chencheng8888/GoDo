package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint      `gorm:"primarykey;column:id"`
	Email     string    `gorm:"column:email;uniqueIndex"`
	UserName  string    `gorm:"column:username;uniqueIndex"`
	Password  string    `gorm:"column:password"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
