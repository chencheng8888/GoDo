package dao

import (
	"github.com/chencheng8888/GoDo/dao/model"
	"gorm.io/gorm"
)

type TaskLogDao struct {
	db *gorm.DB
}

func NewTaskLogDao(db *gorm.DB) *TaskLogDao {
	return &TaskLogDao{db: db}
}

func (t *TaskLogDao) CreateTaskLog(taskLog *model.TaskLog) error {
	return t.db.Create(taskLog).Error
}

func (t *TaskLogDao) FindByUserName(userName string, page, pageSize int) ([]model.TaskLog, int64, error) {
	var (
		logs  []model.TaskLog
		total int64
	)

	offset := (page - 1) * pageSize

	query := t.db.Model(&model.TaskLog{}).
		Where("name = ?", userName)

	// 查询总条数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页 + 按 start_time 逆序排序
	err := query.
		Offset(offset).
		Limit(pageSize).
		Order("start_time DESC").
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
