package dao

import (
	"github.com/chencheng8888/GoDo/model"
	"gorm.io/gorm"
)

type TaskLogDao struct {
	db *gorm.DB
}

func (t *TaskLogDao) CreateTaskLog(taskLog model.TaskLog) error {
	return t.db.Create(&taskLog).Error
}
