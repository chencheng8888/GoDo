package scheduler

import (
	"context"
	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/dao"
	"github.com/chencheng8888/GoDo/model"
	"github.com/chencheng8888/GoDo/pkg/log"
	"github.com/google/wire"
	"github.com/panjf2000/ants/v2"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var (
	ProviderSet = wire.NewSet(NewScheduler, NewLogMiddleware, NewTaskLogMiddleware)
)

type Scheduler struct {
	c        *cron.Cron
	executor Executor

	log         *zap.SugaredLogger
	taskInfoDao *dao.TaskInfoDao

	pool *ants.Pool

	schedulerCtx context.Context

	cancelFunc context.CancelFunc
}

func NewScheduler(conf *config.ScheduleConfig, logMiddleware *LogMiddleware, taskLogMiddleware *TaskLogMiddleware, taskInfoDao *dao.TaskInfoDao, logger *zap.SugaredLogger) (*Scheduler, error) {

	var options []cron.Option

	options = append(options, cron.WithLogger(log.NewCronLogger(logger)))

	if conf.WithSeconds {
		options = append(options, cron.WithSeconds())
	}

	c := cron.New(options...)

	pool, err := ants.NewPool(conf.GoroutinesSize, ants.WithLogger(log.NewAntsLogger(logger)))
	if err != nil {
		return nil, err
	}

	executor := Chain(BaseExecutor, logMiddleware.Handler, taskLogMiddleware.Handler)

	schedulerCtx, cancel := context.WithCancel(context.Background())

	s := &Scheduler{
		c:            c,
		executor:     executor,
		log:          logger,
		taskInfoDao:  taskInfoDao,
		pool:         pool,
		schedulerCtx: schedulerCtx,
		cancelFunc:   cancel,
	}

	// Initialize tasks
	s.initializeTasks()
	return s, nil
}

func (s *Scheduler) AddTask(t Task) (int, error) {
	return s.addTask(t, false)
}

func (s *Scheduler) addTask(t Task, addCronOnly bool) (int, error) {
	id, err := s.c.AddFunc(t.scheduledTime, func() {
		err := s.pool.Submit(func() {
			s.executor(s.schedulerCtx, t)
		})
		if err != nil {
			s.log.Errorf("submit task to pool failed: %s,task:%v", err, t)
		}
	})
	if err != nil {
		return -1, err
	}
	t.id = int(id)

	if !addCronOnly {
		err = s.taskInfoDao.CreateTaskInfo(newModel(t))
		if err != nil {
			// rollback cron job addition
			s.c.Remove(cron.EntryID(id))
			return -1, err
		}
	}
	s.log.Infof("add a task: %+v", t)
	return int(id), nil
}

func (s *Scheduler) ListTasks(userName string) []Task {
	var tasks []Task
	taskInfos, err := s.taskInfoDao.GetTaskInfosByOwnerName(userName)
	if err != nil {
		s.log.Errorf("get task info by owner_name=%s error: %s", userName, err)
		return []Task{}
	}

	for _, taskInfo := range taskInfos {
		task, err := NewTaskFromModel(taskInfo)
		if err != nil {
			s.log.Errorf("new task from model failed: %s", err)
			continue
		}
		tasks = append(tasks, task)
	}
	s.log.Infof("📋 Listed tasks for user:%s,res is [%v]", userName, tasks)
	return tasks
}

func (s *Scheduler) RemoveTask(userName string, taskId int) error {
	err := s.taskInfoDao.DeleteTaskInfoByTaskId(userName, taskId)
	if err != nil {
		s.log.Errorf("delete task info by (user_name=%s and task_id=%d) error: %s", userName, taskId, err)
		return err
	}
	s.c.Remove(cron.EntryID(taskId))
	return nil
}

func (s *Scheduler) Start() {
	s.log.Info("🚩Task scheduler start")
	s.c.Start()
}

func (s *Scheduler) Stop() {
	// 取消上下文
	s.cancelFunc()
	// 停止 cron 调度器
	ctx := s.c.Stop()
	<-ctx.Done()
	// 释放 goroutine 池资源
	s.pool.Release()
	s.log.Info("✔️Task scheduler stopped")
}

func (s *Scheduler) initializeTasks() {
	s.log.Infof("🤖start initialize tasks from db...")

	taskInfos, err := s.taskInfoDao.ListTaskInfo()
	if err != nil {
		s.log.Errorf("initialize tasks:failed to get task info from db: %v", err)
	}
	for _, taskInfo := range taskInfos {
		task, err := NewTaskFromModel(taskInfo)
		if err != nil {
			s.log.Errorf("initialize tasks:failed to create task from model[%v]: %v", taskInfo, err)
			continue
		}
		taskId, err := s.addTask(task, true)
		if err != nil {
			s.log.Errorf("initialize tasks:failed to add task[%v]: %v", task, err)
			continue
		}
		err = s.taskInfoDao.UpdateTaskIdByID(taskInfo.ID, taskId)
		if err != nil {
			s.log.Errorf("initialize tasks:failed to update task info: %v", err)
		}
	}
	s.log.Infof("✅initialize tasks from db finished")
}

func newModel(task Task) *model.TaskInfo {
	return &model.TaskInfo{
		TaskId:        task.id,
		TaskName:      task.taskName,
		OwnerName:     task.ownerName,
		ScheduledTime: task.scheduledTime,
		Description:   task.description,
		JobType:       task.f.Type(),
		Job:           task.f.ToJson(),
	}
}
