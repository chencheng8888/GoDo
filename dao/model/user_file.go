package model

import "time"

type UserFile struct {
	UserName  string    `gorm:"column:user_name;type:varchar(255);primaryKey"`
	FileName  string    `gorm:"column:file_name;type:varchar(255)"`
	Size      int64     `gorm:"column:size"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime"`
}

func (u *UserFile) TableName() string {
	return "user_files"
}
