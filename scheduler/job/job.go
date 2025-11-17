package job

import "context"

type Job interface {
	Run(ctx context.Context)
	Type() string
	Output() <-chan string
	ErrOutput() <-chan string

	ToJson() string
	UnmarshalFromJson(jsonStr string) error
}

const (
	ShellJobType = "shell"
)
