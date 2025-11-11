package scheduler

import (
	"fmt"
	"github.com/chencheng8888/GoDo/model"
	"github.com/chencheng8888/GoDo/scheduler/job"
	"time"
)

type Task struct {
	id            int
	taskName      string // 任务名称
	scheduledTime string // cron表达式
	ownerName     string // 拥有者
	description   string // 描述
	f             job.Job
}

func (t *Task) String() string {
	return fmt.Sprintf("Task{id: %d, taskName: %s, scheduledTime: %s, ownerName: %s, description: %s, job: %s}",
		t.id, t.taskName, t.scheduledTime, t.ownerName, t.description, t.f.ToJson())
}

func NewTask(taskName, ownerName, scheduledTime, description string, job job.Job) Task {
	return Task{
		taskName:      taskName,
		scheduledTime: scheduledTime,
		ownerName:     ownerName,
		description:   description,
		f:             job,
	}
}

func NewTaskFromModel(taskInfo *model.TaskInfo) (Task, error) {
	j, err := GetJob(taskInfo.JobType)
	if err != nil {
		return Task{}, err
	}

	err = j.UnmarshalFromJson(taskInfo.Job)
	if err != nil {
		return Task{}, err
	}

	return Task{
		id:            taskInfo.TaskId,
		taskName:      taskInfo.TaskName,
		scheduledTime: taskInfo.ScheduledTime,
		ownerName:     taskInfo.OwnerName,
		description:   taskInfo.Description,
		f:             j,
	}, nil
}

func GetJob(jobType string) (job.Job, error) {
	switch jobType {
	case job.ShellJobType:
		return new(job.ShellJob), nil
	default:
		return nil, fmt.Errorf("job type unknown")
	}
}

type TaskResult struct {
	StartTime time.Time
	EndTime   time.Time
	Output    string
	ErrOutput string
}
