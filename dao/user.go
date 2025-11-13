package dao

import "gorm.io/gorm"

type UserDao struct {
	db *gorm.DB
}

func (u *UserDao) GetPasswordByUserName(userName string) (string, error) {
	var password string
	err := u.db.Table("users").Where("user_name = ?", userName).Select("password").Scan(&password).Error
	return password, err
}
