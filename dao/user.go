package dao

import "gorm.io/gorm"

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{db: db}
}

func (u *UserDao) GetUser(username string) (string, error) {
	var password string
	err := u.db.Table("users").Where("user_name = ?", username).Select("password").Scan(&password).Error
	return password, err
}
