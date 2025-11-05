package task

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"sync"
)

type Job interface {
	cron.Job
	Content() string
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
}

func (t *Task) Run() {
	t.f.Run()
}

func (t *Task) String() string {
	return fmt.Sprintf("Task{id:%d, name: %s, scheduledTime: %s, owner: %s, description: %s, job: %s}", t.id, t.taskName, t.scheduledTime, t.ownerName, t.description, t.f.Content())
}

type Scheduler struct {
	tasks map[int]*Task
	mu    sync.RWMutex
	c     *cron.Cron
}

func (s *Scheduler) AddTask(t *Task) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
