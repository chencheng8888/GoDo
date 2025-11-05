package job

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type ShellJob struct {
	Command   string      // shell 命令
	Args      []string    // 命令参数
	output    chan string // 标准输出
	errOutput chan string // 错误输出
}

func NewShellJob(command string, args ...string) *ShellJob {
	return &ShellJob{
		Command:   command,
		Args:      args,
		output:    make(chan string, 100),
		errOutput: make(chan string, 100),
	}
}

func (s *ShellJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.Command, s.Args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	// 只写入一次
	if err != nil {
		if stderrStr == "" {
			stderrStr = err.Error()
		}
		s.errOutput <- fmt.Sprintf("command error: %v\n%s", err, stderrStr)
		return
	}

	s.output <- stdoutStr
}

func (s *ShellJob) Content() string {
	return fmt.Sprintf("%s %v", s.Command, s.Args)
}

func (s *ShellJob) Output() <-chan string {
	return s.output
}

func (s *ShellJob) ErrOutput() <-chan string {
	return s.errOutput
}
