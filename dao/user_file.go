package dao

import (
	"errors"
	"github.com/chencheng8888/GoDo/dao/model"
	"gorm.io/gorm"
)

var (
	UserFileNotFoundErr = errors.New("user file record not found")
)

type UserFileDao struct {
	db *gorm.DB
}

func NewUserFileDao(db *gorm.DB) *UserFileDao {
	return &UserFileDao{db: db}
}

func (u *UserFileDao) ListUserFiles(userName string) ([]string, error) {
	var files []string

	err := u.db.Model(&model.UserFile{}).Where("user_name = ?", userName).Pluck("file_name", &files).Error
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (u *UserFileDao) CountFiles(userName string) (int64, error) {
	var cnt int64

	err := u.db.Model(&model.UserFile{}).Where("user_name = ?", userName).Count(&cnt).Error
	if err != nil {
		return 0, err
	}

	return cnt, nil
}

func (u *UserFileDao) AddUserFileRecord(userName string, fileName string, fileSize int64) error {
	return u.db.Model(&model.UserFile{}).Create(&model.UserFile{
		UserName: userName,
		FileName: fileName,
		Size:     fileSize,
	}).Error
}

func (u *UserFileDao) DeleteUserFileRecord(userName string, fileName string) error {
	res := u.db.Model(&model.UserFile{}).Where("user_name = ? AND file_name = ?", userName, fileName).Delete(&model.UserFile{})
	if res.RowsAffected == 0 {
		return UserFileNotFoundErr
	}
	return res.Error
}
