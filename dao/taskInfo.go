package dao

import (
	"fmt"
	"github.com/chencheng8888/GoDo/model"
	"gorm.io/gorm"
)

type TaskInfoDao struct {
	db *gorm.DB
}

func NewTaskInfoDao(db *gorm.DB) *TaskInfoDao {
	return &TaskInfoDao{
		db: db,
	}
}

func (t *TaskInfoDao) CreateTaskInfo(taskInfo *model.TaskInfo) error {
	return t.db.Create(&taskInfo).Error
}

func (t *TaskInfoDao) ListTaskInfo() ([]*model.TaskInfo, error) {
	var taskInfos []*model.TaskInfo
	err := t.db.Find(&taskInfos).Error
	return taskInfos, err
}

func (t *TaskInfoDao) GetTaskInfosByOwnerName(ownerName string) ([]*model.TaskInfo, error) {
	var taskInfos []*model.TaskInfo
	err := t.db.Where("owner_name = ?", ownerName).Find(&taskInfos).Error
	return taskInfos, err
}

func (t *TaskInfoDao) DeleteTaskInfoByTaskId(userName string, taskId int) error {
	res := t.db.Where("owner_name = ? and task_id = ?", userName, taskId).Delete(&model.TaskInfo{})

	if res.RowsAffected == 0 {
		return fmt.Errorf("no task has been found, so can not be delete")
	}
	return res.Error
}

func (t *TaskInfoDao) UpdateTaskIdByID(Id uint, newTaskId int) error {
	return t.db.Model(&model.TaskInfo{}).Where("id = ?", Id).Update("task_id", newTaskId).Error
}
