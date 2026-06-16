package eino

import (
	"context"
	"fmt"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
)

// setupEnv 写入沙箱内 .env 文件，供 eino-agent 与 bash 进程读取模型 / trace 配置。
func (r *Runner) setupEnv(ctx context.Context) error {
	var lines []string

	if r.req.ModelConfig.APIKey != "" {
		lines = append(lines, fmt.Sprintf("OPENAI_API_KEY=%s", r.req.ModelConfig.APIKey))
	}

	// 提取一次 traceparent，复用给 baseURL 拼接与 TRACEPARENT env 行。
	traceParent := ""
	if r.req.TraceContext != nil {
		traceParent = r.req.TraceContext["traceparent"]
	}

	baseURL := r.req.ModelConfig.BaseURL
	// 若存在 traceparent，把 traceId/spanId 编码到 baseURL 路径里。
	// BFF 侧新增带 /trace/:traceId/span/:spanId/ 参数的路由来接收这种请求。
	if traceParent != "" {
		if parts := strings.Split(traceParent, "-"); len(parts) == 4 {
			baseURL = baseURL + "/trace/" + parts[1] + "/span/" + parts[2]
		}
	}
	if baseURL != "" {
		lines = append(lines, fmt.Sprintf("OPENAI_BASE_URL=%s", baseURL))
	}
	if r.req.ModelConfig.Model != "" {
		lines = append(lines, fmt.Sprintf("OPENAI_MODEL_ID=%s", r.req.ModelConfig.Model))
	}

	// 追加 trace 环境变量，供沙箱内 bash 进程（含 curl 调用）继续传播 trace。
	if traceParent != "" {
		lines = append(lines, fmt.Sprintf("TRACEPARENT=%s", traceParent))
	}
	if r.req.TraceContext != nil {
		if ts := r.req.TraceContext["tracestate"]; ts != "" {
			lines = append(lines, fmt.Sprintf("TRACESTATE=%s", ts))
		}
		if bg := r.req.TraceContext["baggage"]; bg != "" {
			lines = append(lines, fmt.Sprintf("BAGGAGE=%s", bg))
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := r.sb.WriteFile(ctx, ".env", []byte(content)); err != nil {
		return fmt.Errorf("failed to create .env: %w", err)
	}
	log.Infof("%s .env file created in sandbox workspace", r.logPrefix)
	return nil
}

// setupWorkspaceDirs 创建 skills/、output/、tmp/ 目录，并把宿主 skills 与 input 复制进沙箱。
// eino-agent HTTP 服务从 workspace/skills/ 加载技能。
func (r *Runner) setupWorkspaceDirs(ctx context.Context) error {
	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", "skills"); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	for _, skill := range r.req.Skills {
		log.Infof("%s copying skill from %s to skills/", r.logPrefix, skill.Dir)
		if err := r.sb.CopyToSandbox(ctx, skill.Dir, "skills"); err != nil {
			return fmt.Errorf("failed to copy skill to workspace: %w", err)
		}
	}

	if r.req.InputDir != "" {
		if err := r.sb.CopyToSandbox(ctx, r.req.InputDir); err != nil {
			return fmt.Errorf("failed to copy input to workspace: %w", err)
		}
	}

	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", "output"); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", "tmp"); err != nil {
		return fmt.Errorf("failed to create tmp directory: %w", err)
	}
	return nil
}
