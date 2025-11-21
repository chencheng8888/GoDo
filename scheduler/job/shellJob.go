package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	ShellJobType = "shell"
)

type ShellJob struct {
	Command   string        `json:"command"`   // shell 命令
	Args      []string      `json:"args"`      // 命令参数
	UseShell  bool          `json:"use_shell"` //是否通过系统默认 Shell 执行 (true: 可以运行内建命令和脚本, false: 直接运行可执行文件)
	Timeout   time.Duration `json:"timeout"`
	output    chan string   // 标准输出
	errOutput chan string   // 错误输出

	workDir string // 工作目录
}

func (s *ShellJob) Content() string {
	if s == nil {
		return ""
	}
	res, _ := json.Marshal(s)
	return string(res)
}

func NewShellJob(useShell bool, timeOut time.Duration, workDir, command string, args ...string) *ShellJob {
	return &ShellJob{
		Command:   command,
		Args:      args,
		UseShell:  useShell,
		Timeout:   timeOut,
		output:    make(chan string, 100),
		errOutput: make(chan string, 100),
		workDir:   workDir,
	}
}

func (s *ShellJob) Type() string {
	return ShellJobType
}

func (s *ShellJob) Run(ctx context.Context) {
	shellCtx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	var cmd *exec.Cmd

	if s.UseShell {
		// --- 明确指定终端/Shell 解释器 ---
		var shell string
		var shellArgs []string

		// 将 Command 和 Args 合并成一个完整的命令字符串
		// 这样可以处理管道、重定向、Shell 变量等复杂的 Shell 语法
		var fullCommand string
		if len(s.Args) > 0 {
			fullCommand = s.Command + " " + strings.Join(s.Args, " ")
		} else {
			fullCommand = s.Command
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

		cmd = exec.CommandContext(shellCtx, shell, shellArgs...)
	} else {
		// --- 直接运行可执行文件 (原有的方式) ---
		cmd = exec.CommandContext(shellCtx, s.Command, s.Args...)
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
		s.errOutput <- fmt.Sprintf("Command error: %v\n%s", err, stderrStr)
		return
	}

	s.output <- stdoutStr
}

func (s *ShellJob) Output() <-chan string {
	return s.output
}

func (s *ShellJob) ErrOutput() <-chan string {
	return s.errOutput
}

func (s *ShellJob) ToJson() string {
	type Alias ShellJob // 防止递归调用
	res, _ := json.Marshal(&struct {
		WorkDir string `json:"work_dir"`
		*Alias
	}{
		WorkDir: s.workDir,
		Alias:   (*Alias)(s),
	})
	return string(res)
}

func (s *ShellJob) UnmarshalFromJson(jsonStr string) error {
	if s == nil {
		return fmt.Errorf("cannot unmarshall from json: nil pointer")
	}

	type Alias ShellJob
	aux := &struct {
		WorkDir string `json:"work_dir"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal([]byte(jsonStr), &aux); err != nil {
		return err
	}
	s.workDir = aux.WorkDir
	s.output = make(chan string, 100)
	s.errOutput = make(chan string, 100)
	return nil
}
