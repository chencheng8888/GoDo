package job

import (
	"github.com/robfig/cron/v3"
)

type Job interface {
	cron.Job
	Type() string
	Output() <-chan string
	ErrOutput() <-chan string

	ToJson() string
	UnmarshalFromJson(jsonStr string) error
}

const (
	ShellJobType = "shell"
)
