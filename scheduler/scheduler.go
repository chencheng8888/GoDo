package scheduler

import (
	"fmt"
	"github.com/chencheng8888/GoDo/config"
	"github.com/robfig/cron/v3"
	"sync"
)

type Scheduler struct {
	tasks    map[int]Task
	mu       sync.RWMutex
	c        *cron.Cron
	executor Executor
}

func NewScheduler(conf *config.ScheduleConfig, logMiddleware *LogMiddleware, taskLogMiddleware *TaskLogMiddleware) *Scheduler {
	var c *cron.Cron
	if conf.WithSeconds {
		c = cron.New(cron.WithSeconds())
	} else {
		c = cron.New()
	}

	executor := Chain(BaseExecutor, logMiddleware.Handler, taskLogMiddleware.Handler)

	return &Scheduler{
		tasks:    make(map[int]Task),
		c:        c,
		executor: executor,
	}
}

func (s *Scheduler) AddTask(t Task) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, err := s.c.AddFunc(t.scheduledTime, func() {
		s.executor(t)
	})
	if err != nil {
		return -1, err
	}
	t.id = int(id)
	s.tasks[int(id)] = t
	return int(id), nil
}

func (s *Scheduler) ListTasks(userName string) []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var tasks []Task
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
		return fmt.Errorf("scheduler with id %d not found", id)
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
