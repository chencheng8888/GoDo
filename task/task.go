package task

import (
	"fmt"
	"github.com/chencheng8888/GoDo/dao"
	"github.com/chencheng8888/GoDo/model"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

type Job interface {
	cron.Job
	Content() string
	Output() <-chan string
	ErrOutput() <-chan string
}

type TaskFunc func()

func (tf TaskFunc) Run() {
	tf()
}

func (tf TaskFunc) Content() string {
	return "this is a go func"
}

type Task struct {
	id            int
	taskName      string // 任务名称
	scheduledTime string // cron表达式
	ownerName     string // 拥有者
	description   string // 描述
	f             Job

	taskLogDao *dao.TaskLogDao
}

func NewTask(taskName, ownerName, scheduledTime, description string, job Job) *Task {
	return &Task{
		taskName:      taskName,
		scheduledTime: scheduledTime,
		ownerName:     ownerName,
		description:   description,
		f:             job,
	}
}

func (t *Task) Run() {
	startTime := time.Now()

	var (
		errOutput string
		isPanic   bool
	)

	func() {
		defer func() {
			if r := recover(); r != nil {
				isPanic = true
				errOutput = fmt.Sprintf("%v", r)
				// fmt.Println("catch panic:", r)
			}
		}()
		t.f.Run()
	}()

	endTime := time.Now()

	taskLog := model.TaskLog{
		TaskId:    t.id,
		Name:      t.taskName,
		Content:   t.f.Content(),
		StartTime: startTime,
		EndTime:   endTime,
	}

	if isPanic {
		taskLog.ErrOutput = fmt.Sprintf("panic occurred here: %v", errOutput)
	} else {
		taskLog.Output = readChannel(t.f.Output())
		taskLog.ErrOutput = readChannel(t.f.ErrOutput())
	}

	err := t.taskLogDao.CreateTaskLog(taskLog)
	if err != nil {
		// TODO: 打日志
	}
	// TODO: 打日志
}

func readChannel(ch <-chan string) string {
	var result string
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return result
			}
			result += msg
		default:
			return result
		}
	}
}

func (t *Task) String() string {
	return fmt.Sprintf("Task{id:%d, name: %s, scheduledTime: %s, owner: %s, description: %s, job: %s}", t.id, t.taskName, t.scheduledTime, t.ownerName, t.description, t.f.Content())
}

type Scheduler struct {
	tasks map[int]*Task
	mu    sync.RWMutex
	c     *cron.Cron

	taskLogDao *dao.TaskLogDao
}

func (s *Scheduler) AddTask(t *Task) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t.taskLogDao = s.taskLogDao

	id, err := s.c.AddJob(t.scheduledTime, t)
	if err != nil {
		return -1, err
	}
	t.id = int(id)
	s.tasks[int(id)] = t
	return int(id), nil
}

func (s *Scheduler) ListTasks(userName string) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var tasks []*Task
	for _, t := range s.tasks {
		if t.ownerName != userName {
			continue
		}

		tasks = append(tasks, t)
	}
	return tasks
}

func (s *Scheduler) RemoveTaskById(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return fmt.Errorf("task with id %d not found", id)
	}
	s.c.Remove(cron.EntryID(id))
	delete(s.tasks, id)
	return nil
}

func (s *Scheduler) Start() {
	s.c.Start()
}

func (s *Scheduler) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
}
