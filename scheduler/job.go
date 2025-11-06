package scheduler

import (
	"github.com/robfig/cron/v3"
)

type Job interface {
	cron.Job
	Content() string
	Output() <-chan string
	ErrOutput() <-chan string
}
