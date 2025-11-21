package domain

import (
	"fmt"
	"github.com/chencheng8888/GoDo/dao/model"
	"time"

	"github.com/chencheng8888/GoDo/scheduler/job"
)

type Task struct {
	id            string
	taskName      string // 任务名称
	scheduledTime string // cron表达式
	ownerName     string // 拥有者
	description   string // 描述
	f             Job
}

func (t *Task) String() string {
	return fmt.Sprintf("Task{id: %v, taskName: %s, scheduledTime: %s, ownerName: %s, description: %s, job: %v}",
		t.id, t.taskName, t.scheduledTime, t.ownerName, t.description, t.f)
}

func NewTask(id, taskName, ownerName, scheduledTime, description string, job Job) Task {
	return Task{
		id:            id,
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

func GetJob(jobType string) (Job, error) {
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

func (t *Task) GetID() string {
	return t.id
}

func (t *Task) GetTaskName() string {
	return t.taskName
}

func (t *Task) GetScheduledTime() string {
	return t.scheduledTime
}

func (t *Task) GetOwnerName() string {
	return t.ownerName
}

func (t *Task) GetDescription() string {
	return t.description
}

func (t *Task) GetJob() Job {
	return t.f
}
