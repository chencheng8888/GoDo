package dao

import (
	"github.com/chencheng8888/GoDo/model"
	"gorm.io/gorm"
)

type TaskLogDao struct {
	db *gorm.DB
}

func NewTaskLogDao(db *gorm.DB) *TaskLogDao {
	return &TaskLogDao{db: db}
}

func (t *TaskLogDao) CreateTaskLog(taskLog model.TaskLog) error {
	return t.db.Create(&taskLog).Error
}
