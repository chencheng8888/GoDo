package model

import (
	"time"
)

type TaskLog struct {
	ID        uint      `gorm:"primarykey"`
	TaskId    int       `gorm:"column:task_id;index"`
	Name      string    `gorm:"type:varchar(100);column:name;index"`   // 任务名称
	Content   string    `gorm:"type:text;column:content"`              // 任务内容，比如 shell 命令或者 Go 函数描述
	Output    string    `gorm:"type:text;column:output"`               // 任务执行输出
	ErrOutput string    `gorm:"type:text;column:err_output"`           // 任务执行错误输出
	StartTime time.Time `gorm:"type:datetime;column:start_time;index"` // 任务开始时间
	EndTime   time.Time `gorm:"type:datetime;column:end_time;index"`   // 任务结束时间
}

func (t *TaskLog) TableName() string {
	return "task_logs"
}
