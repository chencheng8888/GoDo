package dao

import (
	"errors"

	"gorm.io/gorm"
)

type UserDao struct {
	db *gorm.DB
}

// NewUserDao 创建UserDao实例
func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{
		db: db,
	}
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func (u *UserDao) GetPasswordByUserName(userName string) (string, error) {
	var password string
	err := u.db.Table("users").Where("username = ?", userName).Select("password").Scan(&password).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	// 如果密码为空字符串，也表示用户不存在
	if password == "" {
		return "", ErrUserNotFound
	}

	return password, nil
}
