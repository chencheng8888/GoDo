package implement

import (
	"context"
	"errors"
	"fmt"
	"github.com/chencheng8888/GoDo/dao/model"
	"sync"

	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/dao"
	"github.com/chencheng8888/GoDo/pkg/log"
	"github.com/chencheng8888/GoDo/scheduler/domain"
	"github.com/panjf2000/ants/v2"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CronJobFunc func()

func (c CronJobFunc) Run() {
	c()
}

type CronScheduler struct {
	c      *cron.Cron
	parser cron.Parser

	// 维护 业务ID -> 运行时EntryID 的映射
	mapping map[string]cron.EntryID

	mu sync.Mutex

	executor domain.Executor

	log         *zap.SugaredLogger
	taskInfoDao *dao.TaskInfoDao

	pool *ants.Pool

	schedulerCtx context.Context

	cancelFunc context.CancelFunc
}

func NewCronScheduler(conf *config.ScheduleConfig, logMiddleware *domain.LogMiddleware, taskLogMiddleware *domain.TaskLogMiddleware, taskInfoDao *dao.TaskInfoDao, logger *zap.SugaredLogger) (*CronScheduler, error) {

	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	if conf.WithSeconds {
		parser = cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		)
	}

	c := cron.New()

	pool, err := ants.NewPool(conf.GoroutinesSize, ants.WithLogger(log.NewAntsLogger(logger)))
	if err != nil {
		return nil, err
	}

	executor := domain.Chain(domain.BaseExecutor, logMiddleware.Handler, taskLogMiddleware.Handler)

	schedulerCtx, cancel := context.WithCancel(context.Background())

	s := &CronScheduler{
		c:            c,
		parser:       parser,
		mapping:      make(map[string]cron.EntryID),
		executor:     executor,
		log:          logger,
		taskInfoDao:  taskInfoDao,
		pool:         pool,
		schedulerCtx: schedulerCtx,
		cancelFunc:   cancel,
	}

	return s, nil
}

func (s *CronScheduler) AddTask(t domain.Task) error {
	return s.addTask(t, false)
}

func (s *CronScheduler) addTask(t domain.Task, addCronOnly bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查看下mapping中是否有对应id
	if _, exists := s.mapping[t.GetID()]; exists {
		return fmt.Errorf("task id %s already exists", t.GetID())
	}

	// 解析时间是否正确
	sche, err := s.parser.Parse(t.GetScheduledTime())
	if err != nil {
		return fmt.Errorf("parse scheduled_time failed: %s", err)
	}

	if !addCronOnly {
		// 添加数据库失败
		err = s.taskInfoDao.CreateTaskInfo(newModel(t))
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("task id %s already exists in db", t.GetID())
			}
			return err
		}
	}

	id := s.c.Schedule(sche, CronJobFunc(func() {
		err := s.pool.Submit(func() {
			s.executor(s.schedulerCtx, t)
		})
		if err != nil {
			s.log.Errorf("submit task to pool failed: %s,task:%v", err, t)
		}
	}))

	// 更新mapping
	s.mapping[t.GetID()] = id

	s.log.Infof("add a task successfully: %+v", t)
	return nil
}

func (s *CronScheduler) ListTasks(userName string) []domain.Task {
	var tasks []domain.Task
	taskInfos, err := s.taskInfoDao.GetTaskInfosByOwnerName(userName)
	if err != nil {
		s.log.Errorf("get task info by owner_name=%s error: %s", userName, err)
		return []domain.Task{}
	}

	for _, taskInfo := range taskInfos {
		task, err := domain.NewTaskFromModel(taskInfo)
		if err != nil {
			s.log.Errorf("new task from model failed: %s", err)
			continue
		}
		tasks = append(tasks, task)
	}
	s.log.Infof("📋 Listed tasks for user:%s,res is [%v]", userName, tasks)
	return tasks
}

func (s *CronScheduler) RemoveTask(userName string, taskId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cronId, ok := s.mapping[taskId]
	if !ok {
		return fmt.Errorf("task id %s not found", taskId)
	}

	err := s.taskInfoDao.DeleteTaskInfoByTaskId(userName, taskId)
	if err != nil {
		s.log.Errorf("delete task info by (user_name=%s and task_id=%s) error: %s", userName, taskId, err)
		return err
	}

	s.c.Remove(cronId)
	delete(s.mapping, taskId)

	s.log.Infof("delete a task: user_name=%s, task_id=%s", userName, taskId)
	return nil
}

func (s *CronScheduler) Start() {
	s.log.Info("🚩Task scheduler start")
	s.c.Start()
}

func (s *CronScheduler) Stop() {
	// 取消上下文
	s.cancelFunc()
	// 停止 cron 调度器
	ctx := s.c.Stop()
	<-ctx.Done()
	// 释放 goroutine 池资源
	s.pool.Release()
	s.log.Info("✔️Task scheduler stopped")
}

func (s *CronScheduler) InitializeTasks() {
	s.log.Infof("🤖start initialize tasks from db...")

	taskInfos, err := s.taskInfoDao.ListTaskInfo()
	if err != nil {
		s.log.Errorf("initialize tasks:failed to get task info from db: %v", err)
	}
	for _, taskInfo := range taskInfos {
		task, err := domain.NewTaskFromModel(taskInfo)
		if err != nil {
			s.log.Errorf("initialize tasks:failed to new task from model[%v]: %v", taskInfo, err)
			continue
		}
		err = s.addTask(task, true)
		if err != nil {
			s.log.Errorf("initialize tasks:failed to add task[%v]: %v", taskInfo, err)
			continue
		}
	}
	s.log.Infof("✅initialize tasks from db finished")
}

func newModel(task domain.Task) *model.TaskInfo {
	return &model.TaskInfo{
		TaskId:        task.GetID(),
		TaskName:      task.GetTaskName(),
		OwnerName:     task.GetOwnerName(),
		ScheduledTime: task.GetScheduledTime(),
		Description:   task.GetDescription(),
		JobType:       task.GetJob().Type(),
		Job:           task.GetJob().ToJson(),
	}
}
