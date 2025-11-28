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

func (t *TaskLogDao) FindByUserName(ownerName string, page, pageSize int) ([]model.TaskLog, int64, error) {
	var (
		logs  []model.TaskLog
		total int64
	)

	offset := (page - 1) * pageSize

	// 构建 JOIN 查询（不加载 TaskInfo，只 join 获取 owner_name 筛选条件）
	query := t.db.Table("task_logs tl").
		Joins("JOIN task_infos ti ON tl.task_id = ti.task_id").
		Where("ti.owner_name = ?", ownerName)

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询 + 排序
	err := query.
		Order("tl.start_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
