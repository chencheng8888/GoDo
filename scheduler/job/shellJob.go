package job

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type ShellJob struct {
	command   string   // shell 命令
	args      []string // 命令参数
	useShell  bool     //是否通过系统默认 Shell 执行 (true: 可以运行内建命令和脚本, false: 直接运行可执行文件)
	timeout   time.Duration
	output    chan string // 标准输出
	errOutput chan string // 错误输出

	workDir string // 工作目录
}

func NewShellJob(useShell bool, timeOut time.Duration, workDir, command string, args ...string) *ShellJob {
	return &ShellJob{
		command:   command,
		args:      args,
		useShell:  useShell,
		timeout:   timeOut,
		output:    make(chan string, 100),
		errOutput: make(chan string, 100),
		workDir:   workDir,
	}
}

func (s *ShellJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	var cmd *exec.Cmd

	if s.useShell {
		// --- 明确指定终端/Shell 解释器 ---
		var shell string
		var shellArgs []string

		// 将 command 和 args 合并成一个完整的命令字符串
		// 这样可以处理管道、重定向、Shell 变量等复杂的 Shell 语法
		var fullCommand string
		if len(s.args) > 0 {
			fullCommand = s.command + " " + strings.Join(s.args, " ")
		} else {
			fullCommand = s.command
		}

		if runtime.GOOS == "windows" {
			// Windows: 使用 cmd.exe /C 执行命令
			shell = "cmd"
			shellArgs = []string{"/C", fullCommand}
		} else {
			// Linux/macOS: 使用 /bin/bash -c 执行命令
			shell = "/bin/bash"
			shellArgs = []string{"-c", fullCommand}
		}

		cmd = exec.CommandContext(ctx, shell, shellArgs...)
	} else {
		// --- 直接运行可执行文件 (原有的方式) ---
		cmd = exec.CommandContext(ctx, s.command, s.args...)
	}

	if len(s.workDir) > 0 {
		cmd.Dir = s.workDir
	}

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
	return fmt.Sprintf("shell job: [command:%s, args:%v, useShell:%v, timeOut:%v]", s.command, s.args, s.useShell, s.timeout)
}

func (s *ShellJob) Output() <-chan string {
	return s.output
}

func (s *ShellJob) ErrOutput() <-chan string {
	return s.errOutput
}
