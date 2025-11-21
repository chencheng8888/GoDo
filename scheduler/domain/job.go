package domain

import "context"

type Job interface {
	Run(ctx context.Context)
	Type() string
	Content() string

	Output() <-chan string
	ErrOutput() <-chan string

	ToJson() string
	UnmarshalFromJson(jsonStr string) error
}
