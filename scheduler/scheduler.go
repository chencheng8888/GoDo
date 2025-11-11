package scheduler

import (
	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/dao"
	"github.com/chencheng8888/GoDo/model"
	"github.com/google/wire"
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
}

func NewScheduler(conf *config.ScheduleConfig, logMiddleware *LogMiddleware, taskLogMiddleware *TaskLogMiddleware, taskInfoDao *dao.TaskInfoDao, log *zap.SugaredLogger) *Scheduler {
	var c *cron.Cron
	if conf.WithSeconds {
		c = cron.New(cron.WithSeconds())
	} else {
		c = cron.New()
	}

	executor := Chain(BaseExecutor, logMiddleware.Handler, taskLogMiddleware.Handler)

	s := &Scheduler{
		c:           c,
		executor:    executor,
		log:         log,
		taskInfoDao: taskInfoDao,
	}

	// Initialize tasks
	s.initializeTasks()
	return s
}

func (s *Scheduler) AddTask(t Task) (int, error) {
	return s.addTask(t, false)
}

func (s *Scheduler) addTask(t Task, addCronOnly bool) (int, error) {
	id, err := s.c.AddFunc(t.scheduledTime, func() {
		s.executor(t)
	})
	if err != nil {
		return -1, err
	}
	t.id = int(id)

	if !addCronOnly {
		err = s.taskInfoDao.CreateTaskInfo(newModel(t))
		if err != nil {
			// rollback cron job addition
			_ = s.removeTaskByIds(int(id), true)
			return -1, err
		}
	}

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

func (s *Scheduler) RemoveTaskById(id int) error {
	return s.removeTaskByIds(id, false)
}

func (s *Scheduler) removeTaskByIds(id int, removeCronOnly bool) error {
	if !removeCronOnly {
		err := s.taskInfoDao.DeleteTaskInfoByTaskId(id)
		if err != nil {
			s.log.Errorf("delete task info by task_id=%d error: %s", id, err)
			return err
		}
	}
	s.c.Remove(cron.EntryID(id))
	return nil
}

func (s *Scheduler) Start() {
	s.log.Info("🚩Task scheduler start")
	s.c.Start()
}

func (s *Scheduler) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
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
