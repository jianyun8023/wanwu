package shared

import (
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// FinishReasonStop 是 OpenAI / eino 约定的正常停止 FinishReason 值。
// 上游 internal/agent-service/pkg/util.StopMessage 依赖此值判定沙箱收尾。
const FinishReasonStop = "stop"

// FinalErrorSource 标记兜底消息的产生层，作为前缀写入 content：
// "error[<source>]: <err.Error()>"。调用方可用 strings.HasPrefix
// 分别识别 agent 业务错误与 runner 通道错误。
type FinalErrorSource string

const (
	// FinalErrorSourceAgent 表示沙箱容器内 eino-agent 自报的错误
	// （LLM 调用失败、初始化失败、handler panic、ctx 超时等）。
	FinalErrorSourceAgent FinalErrorSource = "agent"
	// FinalErrorSourceRunner 表示宿主机 wga-sandbox runner 观察到的
	// 通道/容器异常（unexpected EOF、connection reset、i/o timeout 等）。
	FinalErrorSourceRunner FinalErrorSource = "runner"
)

// BuildFinalAssistantMessage 构造 Role=Assistant + FinishReason=stop 的兜底消息。
// err != nil 时 content 形如 "error[agent]: <err>"；err == nil 时 content 为空字符串。
func BuildFinalAssistantMessage(source FinalErrorSource, err error) *schema.Message {
	content := ""
	if err != nil {
		content = fmt.Sprintf("error[%s]: %s", source, err.Error())
	}
	return &schema.Message{
		Role:    schema.Assistant,
		Content: content,
		ResponseMeta: &schema.ResponseMeta{
			FinishReason: FinishReasonStop,
		},
	}
}

// BuildFinalAgentEvent 把兜底 assistant+stop 消息包成 AgentEvent，
// 便于 server 层通过 SSEWriter.WriteAgentEvent 直接写出。
func BuildFinalAgentEvent(source FinalErrorSource, err error) *adk.AgentEvent {
	return &adk.AgentEvent{
		Output: &adk.AgentOutput{
			MessageOutput: &adk.MessageVariant{
				IsStreaming: false,
				Message:     BuildFinalAssistantMessage(source, err),
				Role:        schema.Assistant,
			},
		},
	}
}

// IsFinalStopMessage 判断 schema.Message 是否已经是 assistant+stop 的收尾消息。
func IsFinalStopMessage(m *schema.Message) bool {
	if m == nil {
		return false
	}
	if m.Role != schema.Assistant {
		return false
	}
	if m.ResponseMeta == nil {
		return false
	}
	return m.ResponseMeta.FinishReason == FinishReasonStop
}
