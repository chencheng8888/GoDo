package scheduler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// 工作目录应该在项目根目录下运行测试，以确保命令能够正确执行。
func TestShellJob_Run(t *testing.T) {
	type fields struct {
		command   string
		args      []string
		useShell  bool
		timeOut   time.Duration
		output    chan string
		errOutput chan string
	}
	tests := []struct {
		name        string
		fields      fields
		shouldError bool
	}{
		{
			name: "Test Echo Command",
			fields: fields{
				command:   "echo",
				args:      []string{"Hello, World!"},
				useShell:  true,
				timeOut:   5 * time.Second,
				output:    make(chan string, 100),
				errOutput: make(chan string, 100),
			},
			shouldError: false,
		},
		{
			name: "Test Go Command",
			fields: fields{
				command:   "go",
				args:      []string{"run", "./test/test.go"},
				useShell:  true,
				timeOut:   5 * time.Second,
				output:    make(chan string, 100),
				errOutput: make(chan string, 100),
			},
			shouldError: false,
		},
		{
			name: "Test TimeOut Command",
			fields: fields{
				command:   "go",
				args:      []string{"run", "./test/test.go"},
				useShell:  true,
				timeOut:   2 * time.Second,
				output:    make(chan string, 100),
				errOutput: make(chan string, 100),
			},
			shouldError: true,
		},
		{
			name: "Test Non-existent Command",
			fields: fields{
				command:   "nonexistentcommand",
				args:      []string{},
				useShell:  true,
				timeOut:   5 * time.Second,
				output:    make(chan string, 100),
				errOutput: make(chan string, 100),
			},
			shouldError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShellJob{
				Command:   tt.fields.command,
				Args:      tt.fields.args,
				UseShell:  tt.fields.useShell,
				Timeout:   tt.fields.timeOut,
				output:    tt.fields.output,
				errOutput: tt.fields.errOutput,
			}
			s.Run(context.Background())

			if tt.shouldError {
				// 预期错误，应该从 ErrOutput 收到数据
				select {
				case errOutput := <-s.ErrOutput():
					t.Logf("Received expected error: %s", errOutput)
					assert.Contains(t, errOutput, "error", "Error output should contain the word 'error'")
				case <-s.Output():
					t.Errorf("Expected error output, but received standard output")
				}
			} else {
				// 预期成功，应该从 Output 收到数据
				select {
				case out := <-s.Output():
					t.Logf("Received expected output: %s", out)
				case errOutput := <-s.ErrOutput():
					// 如果收到了错误，则测试失败
					t.Fatalf("Unexpected Command error: %s", errOutput)
				}
			}
		})
	}
}
