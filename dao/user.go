package dao

import (
	"github.com/chencheng8888/GoDo/dao/model"
	"gorm.io/gorm"
)

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{db: db}
}

func (u *UserDao) GetUser(username string) (model.User, error) {
	var user model.User
	err := u.db.Model(&model.User{}).Where("user_name = ?", username).First(&user).Error
	return user, err
}
