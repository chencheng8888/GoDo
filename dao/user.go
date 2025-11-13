package dao

import (
	"errors"

	"github.com/chencheng8888/GoDo/model"
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
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrEmailExists     = errors.New("email already exists")
	ErrUserNameExists  = errors.New("username already exists")
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

// CheckUserExistsByUserName 检查用户名是否已存在
func (u *UserDao) CheckUserExistsByUserName(userName string) (bool, error) {
	var count int64
	err := u.db.Table("users").Where("username = ?", userName).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckUserExistsByEmail 检查邮箱是否已存在
func (u *UserDao) CheckUserExistsByEmail(email string) (bool, error) {
	var count int64
	err := u.db.Table("users").Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateUser 创建新用户
func (u *UserDao) CreateUser(user *model.User) error {
	return u.db.Create(user).Error
}
