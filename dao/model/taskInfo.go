package model

import (
	"gorm.io/gorm"
	"time"
)

type TaskInfo struct {
	ID            uint      `gorm:"primarykey"`
	TaskId        string    `gorm:"column:task_id;type:varchar(255);uniqueIndex"`
	TaskName      string    `gorm:"column:task_name;index"`
	ScheduledTime string    `gorm:"column:scheduled_time"`
	OwnerName     string    `gorm:"column:owner_name"`
	Description   string    `gorm:"column:description"`
	JobType       string    `gorm:"column:job_type"`
	Job           string    `gorm:"column:job;"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (t *TaskInfo) TableName() string {
	return "task_infos"
}

func (t *TaskInfo) BeforeCreate(tx *gorm.DB) error {
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}

func (t *TaskInfo) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}
