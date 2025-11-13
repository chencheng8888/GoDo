package dao

import "gorm.io/gorm"

type UserDao struct {
	db *gorm.DB
}

// NewUserDao 创建UserDao实例
func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{
		db: db,
	}
}

func (u *UserDao) GetPasswordByUserName(userName string) (string, error) {
	var password string
	err := u.db.Table("users").Where("user_name = ?", userName).Select("password").Scan(&password).Error
	return password, err
}
