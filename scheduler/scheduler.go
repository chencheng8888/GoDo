package scheduler

import (
	"github.com/chencheng8888/GoDo/scheduler/domain"
	"github.com/chencheng8888/GoDo/scheduler/implement"
	"github.com/google/wire"
)

var (
	ProviderSet = wire.NewSet(implement.NewCronScheduler, domain.NewLogMiddleware, domain.NewTaskLogMiddleware, NewScheduler)
)

type Scheduler interface {
	AddTask(t domain.Task) error
	ListTasks(userName string) []domain.Task
	RemoveTask(userName string, taskId string) error
	Start()
	Stop()
	InitializeTasks()
}

func NewScheduler(cronScheduler *implement.CronScheduler) Scheduler {
	return cronScheduler
}
