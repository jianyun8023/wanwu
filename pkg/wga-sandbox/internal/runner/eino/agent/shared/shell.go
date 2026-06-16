package shared

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk/filesystem"
)

const (
	maxOutputSize  = 1024 * 1024 // 1MB
	commandTimeout = 5 * time.Minute
)

// ShellOnlyBackend 仅提供 shell 命令执行能力。
// 不实现 filesystem.Backend 的其他文件操作方法，因为 bash 工具只调用 Execute。
type ShellOnlyBackend struct {
	maxOutputSize  int
	commandTimeout time.Duration
	workDir        string
}

func NewShellOnlyBackend(workDir string) *ShellOnlyBackend {
	return &ShellOnlyBackend{
		maxOutputSize:  maxOutputSize,
		commandTimeout: commandTimeout,
		workDir:        workDir,
	}
}

// --- 安全规则 ---

var dangerousPatterns = []*regexp.Regexp{
	// 系统级破坏命令
	regexp.MustCompile(`(?i)\brm\s+(-[a-z]*f|-[a-z]*r|--force|--recursive)\b`),
	regexp.MustCompile(`(?i)\bmkfs\b`),
	regexp.MustCompile(`(?i)\bdd\s+if=`),
	regexp.MustCompile(`(?i)\bchmod\s+777\b`),
	regexp.MustCompile(`(?i)\b(shutdown|reboot|init\s+[06])\b`),
	regexp.MustCompile(`(?i)>\s*/dev/sd`),
	regexp.MustCompile(`(?i)\bcurl\b.*\|\s*(ba)?sh`),
	regexp.MustCompile(`(?i)\bwget\b.*\|\s*(ba)?sh`),
	regexp.MustCompile(`(?i)\bnc\s+-[a-z]*l`),
	regexp.MustCompile(`(?i)/dev/(tcp|udp)/`),
	// 通用代码注入（仅拦截明显危险的）
	regexp.MustCompile(`(?i)\b__import__\s*\(\s*['"]os['"]\s*\)`),
}

var sensitivePathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`/etc/(passwd|shadow|sudoers|hosts)`),
	regexp.MustCompile(`/proc/`),
	regexp.MustCompile(`/sys/`),
	regexp.MustCompile(`/dev/`),
	regexp.MustCompile(`~/`),
	regexp.MustCompile(`/var/lib/(mysql|postgresql)`),
}

var (
	allowedWorkspacePattern = regexp.MustCompile(`/home/root/workspace/[a-zA-Z0-9_-]+/workspace`)
	pathTraversalPattern    = regexp.MustCompile(`\.\./`)
	symlinkPattern          = regexp.MustCompile(`(?i)\bln\s+-[a-z]*s`)
	homeDirPattern          = regexp.MustCompile(`/home/[^/]+`)
)

func validateCommand(command string) error {
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(command) {
			matched := pattern.FindString(command)
			log.Printf("[安全拦截] 检测到危险操作: %s", matched)
			return fmt.Errorf("安全拦截：检测到高危命令片段 [%s]，已拒绝执行", matched)
		}
	}

	// 检查是否访问允许的工作目录
	if !allowedWorkspacePattern.MatchString(command) {
		// 不在允许的工作目录内，检查敏感路径
		for _, pattern := range sensitivePathPatterns {
			if matched := pattern.FindString(command); matched != "" {
				return fmt.Errorf("安全拦截：检测到对敏感路径 [%s] 的访问尝试，仅允许操作工作目录", matched)
			}
		}

		// 检查 /root/ 和 /home/ 路径
		if strings.Contains(command, "/root/") {
			return fmt.Errorf("安全拦截：检测到对敏感路径 [/root/] 的访问尝试，仅允许操作工作目录")
		}
		if matched := homeDirPattern.FindString(command); matched != "" {
			return fmt.Errorf("安全拦截：检测到对敏感路径 [%s] 的访问尝试，仅允许操作工作目录", matched)
		}
	}

	if strings.Contains(command, "..") && pathTraversalPattern.MatchString(command) {
		return fmt.Errorf("安全拦截：检测到路径穿越尝试（../），已拒绝执行")
	}

	if symlinkPattern.MatchString(command) {
		for _, pattern := range sensitivePathPatterns {
			if matched := pattern.FindString(command); matched != "" {
				return fmt.Errorf("安全拦截：禁止创建指向敏感路径 [%s] 的符号链接", matched)
			}
		}
	}

	return nil
}

// Execute 执行 shell 命令，附带安全校验、超时、输出截断与退出码处理。
func (b *ShellOnlyBackend) Execute(ctx context.Context, req *filesystem.ExecuteRequest) (*filesystem.ExecuteResponse, error) {
	if err := validateCommand(req.Command); err != nil {
		exitCode := 1
		return &filesystem.ExecuteResponse{
			Output:   err.Error(),
			ExitCode: &exitCode,
		}, nil
	}

	execCtx, cancel := context.WithTimeout(ctx, b.commandTimeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/C", req.Command)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", req.Command)
	}

	if b.workDir != "" {
		if absWorkDir, err := filepath.Abs(b.workDir); err == nil {
			cmd.Dir = absWorkDir
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if execCtx.Err() == context.DeadlineExceeded {
		exitCode := 124
		return &filesystem.ExecuteResponse{
			Output:   fmt.Sprintf("命令执行超时（限制 %v），已终止。请考虑拆分任务或优化命令。", b.commandTimeout),
			ExitCode: &exitCode,
		}, nil
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		if len(output) > 0 {
			output += "\n"
		}
		output += stderr.String()
	}

	truncated := false
	if len(output) > b.maxOutputSize {
		output = output[:b.maxOutputSize]
		truncated = true
	}

	exitCode := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			return nil, err
		}
		exitCode = exitErr.ExitCode()
	}

	return &filesystem.ExecuteResponse{
		Output:    output,
		ExitCode:  &exitCode,
		Truncated: truncated,
	}, nil
}
