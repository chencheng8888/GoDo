package scheduler

import (
	"context"
	"github.com/google/wire"
)

var (
	ProviderSet = wire.NewSet(NewCronScheduler, NewLogMiddleware, NewTaskLogMiddleware, NewScheduler)
)

type Scheduler interface {
	AddTask(t Task) error
	ListTasks(userName string) []Task
	RemoveTask(userName string, taskId string) error
	Start()
	Stop()
	InitializeTasks()
	RunTask(ctx context.Context, task Task)
}

func NewScheduler(cronScheduler *CronScheduler) Scheduler {
	return cronScheduler
}
