package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Executor func(ctx context.Context, t Task) TaskResult

func BaseExecutor(ctx context.Context, t Task) TaskResult {
	start := time.Now()

	var (
		panicMsg = ""
	)

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicMsg = fmt.Sprintf("%v", r)
			}
		}()
		t.f.Run(ctx)
	}()

	errOutput := readChannel(t.f.ErrOutput())

	if panicMsg != "" {
		panicMsg = "Panic occurred: " + panicMsg
		errOutput = strings.Join([]string{panicMsg, errOutput}, ";")
	}

	return TaskResult{
		StartTime: start,
		EndTime:   time.Now(),
		Output:    readChannel(t.f.Output()),
		ErrOutput: errOutput,
	}
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

func Chain(executor Executor, middlewares ...Middleware) Executor {
	// 从最后一个中间件开始，逐步将 executor 包装起来
	for i := len(middlewares) - 1; i >= 0; i-- {
		executor = middlewares[i](executor)
	}
	// 最终调用最外层的 executor
	return executor
}
