package scheduler

import (
	"time"
)

type Task struct {
	id            int
	taskName      string // 任务名称
	scheduledTime string // cron表达式
	ownerName     string // 拥有者
	description   string // 描述
	f             Job
}

func NewTask(taskName, ownerName, scheduledTime, description string, job Job) Task {
	return Task{
		taskName:      taskName,
		scheduledTime: scheduledTime,
		ownerName:     ownerName,
		description:   description,
		f:             job,
	}
}

type TaskResult struct {
	StartTime time.Time
	EndTime   time.Time
	Output    string
	ErrOutput string
}
