package scheduler

import (
	"fmt"
	"strings"
	"time"
)

type Executor func(t Task) TaskResult

func BaseExecutor(t Task) TaskResult {
	start := time.Now()

	var (
		panicMsg = "no panic occurred"
	)

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicMsg = fmt.Sprintf("%v", r)
			}
		}()
		t.f.Run()
	}()
	return TaskResult{
		StartTime: start,
		EndTime:   time.Now(),
		Output:    readChannel(t.f.Output()),
		ErrOutput: strings.Join([]string{panicMsg, readChannel(t.f.ErrOutput())}, ";"),
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
