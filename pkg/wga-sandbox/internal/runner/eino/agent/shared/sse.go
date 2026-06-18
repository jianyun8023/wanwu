package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudwego/eino/adk"
)

type SSEWriter interface {
	WriteAgentEvent(event *adk.AgentEvent)
}

// sseDebugEnabled 控制 marshalAgentEvent 是否输出完整 payload。
// 默认仅打长度摘要；设置 SSE_DEBUG=1（或任意非空值）启用 full payload。
var sseDebugEnabled = os.Getenv("SSE_DEBUG") != ""

// httpSSEWriter 把 AgentEvent 序列化后以 `data: ...\n\n` 形式刷出。
type httpSSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewHTTPSSEWriter(w http.ResponseWriter, flusher http.Flusher) SSEWriter {
	return &httpSSEWriter{w: w, flusher: flusher}
}

func (h *httpSSEWriter) WriteAgentEvent(event *adk.AgentEvent) {
	jsonData, _ := marshalAgentEvent(event)
	_, _ = fmt.Fprintf(h.w, "data: %s\n\n", jsonData)
	h.flusher.Flush()
}

func marshalAgentEvent(event *adk.AgentEvent) ([]byte, error) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("[SSE] marshal AgentEvent failed: %v", err)
		errEvent := &adk.AgentEvent{Err: fmt.Errorf("内部序列化错误")}
		return json.Marshal(errEvent)
	}
	role := eventRole(event)
	if sseDebugEnabled {
		log.Printf("[SSE] >>> role=%s bytes=%d payload=%s", role, len(jsonData), string(jsonData))
	} else {
		log.Printf("[SSE] >>> role=%s bytes=%d", role, len(jsonData))
	}
	return jsonData, nil
}

func eventRole(event *adk.AgentEvent) string {
	if event == nil || event.Output == nil || event.Output.MessageOutput == nil {
		return ""
	}
	return string(event.Output.MessageOutput.Role)
}

// --- 公共事件处理 ---

// ProcessEvents 消费 adk 事件迭代器并经 SSEWriter 写出。
//
// 返回 sentFinal 表示已经写出一条 Role=Assistant + FinishReason=stop 的兜底/正常收尾消息，
// 调用方据此决定是否还需要补发兜底，避免重复。
//
// 任何 iter.Err 或 forwardMessageEvent 失败（通常是 MessageVariant.GetMessage
// 物化流式消息时出错）都会：
//  1. 先写出原始诊断 event（保留 Err 字段供日志/审计）；
//  2. 紧接着写一条 BuildFinalAgentEvent 兜底 assistant+stop 消息；
//  3. 直接 return，结束事件循环。
func ProcessEvents(ctx context.Context, iter *adk.AsyncIterator[*adk.AgentEvent], w SSEWriter) (eventCount int, sentFinal bool) {
	writeFinal := func(err error) {
		w.WriteAgentEvent(BuildFinalAgentEvent(FinalErrorSourceAgent, err))
		sentFinal = true
	}

	for {
		event, ok := iter.Next()
		if !ok {
			log.Printf("[Events] iterator closed, total=%d", eventCount)
			if !sentFinal && ctx != nil && ctx.Err() != nil {
				// 迭代器自然结束但 ctx 已 cancel/timeout：补一条兜底，便于上游识别。
				writeFinal(ctx.Err())
			}
			return
		}

		if event.Err != nil {
			log.Printf("[Events] error event #%d agent=%s role=%s tool=%s err=%v",
				eventCount, event.AgentName, eventRole(event), eventToolName(event), event.Err)
			// 返回原诊断错误事件给客户端（保持现有逻辑），再补一条兜底 assistant+stop。
			w.WriteAgentEvent(event)
			writeFinal(event.Err)
			return
		}

		eventCount++

		if event.Output == nil || event.Output.MessageOutput == nil {
			log.Printf("[Events] event #%d empty output, skipping", eventCount)
			continue
		}

		if err := forwardMessageEvent(event, w, eventCount); err != nil {
			writeFinal(err)
			return
		}
	}
}

// forwardMessageEvent 把一个带 MessageOutput 的 event 物化为非流式后转发。
// 流式与非流式共用同一条转发路径：MessageVariant.GetMessage() 内部已经处理
// schema.ConcatMessageStream，无需我们手动 Recv 循环。
func forwardMessageEvent(event *adk.AgentEvent, w SSEWriter, eventNum int) error {
	mv := event.Output.MessageOutput
	msg, err := mv.GetMessage()
	if err != nil {
		log.Printf("[Events] event #%d get message failed: agent=%s role=%s tool=%s err=%v",
			eventNum, event.AgentName, string(mv.Role), mv.ToolName, err)
		w.WriteAgentEvent(buildDiagnosticEvent(event, fmt.Errorf("failed to materialize message: %w", err)))
		return err
	}
	if msg == nil {
		log.Printf("[Events] event #%d materialized message is nil: agent=%s role=%s tool=%s",
			eventNum, event.AgentName, string(mv.Role), mv.ToolName)
		err := fmt.Errorf("materialized message is nil")
		w.WriteAgentEvent(buildDiagnosticEvent(event, err))
		return err
	}

	log.Printf("[Events] event #%d role=%s tool=%s streaming=%v content=%d tool_calls=%d",
		eventNum, string(mv.Role), mv.ToolName, mv.IsStreaming, len(msg.Content), len(msg.ToolCalls))

	w.WriteAgentEvent(&adk.AgentEvent{
		AgentName: event.AgentName,
		RunPath:   event.RunPath,
		Output: &adk.AgentOutput{
			MessageOutput: &adk.MessageVariant{
				IsStreaming: false,
				Message:     msg,
				Role:        mv.Role,
				ToolName:    mv.ToolName,
			},
		},
		Action: event.Action,
	})
	return nil
}

func buildDiagnosticEvent(src *adk.AgentEvent, err error) *adk.AgentEvent {
	return &adk.AgentEvent{
		AgentName: src.AgentName,
		RunPath:   src.RunPath,
		Output: &adk.AgentOutput{
			MessageOutput: &adk.MessageVariant{
				IsStreaming: false,
				Role:        src.Output.MessageOutput.Role,
				ToolName:    src.Output.MessageOutput.ToolName,
			},
		},
		Err: err,
	}
}

func eventToolName(event *adk.AgentEvent) string {
	if event == nil || event.Output == nil || event.Output.MessageOutput == nil {
		return ""
	}
	return event.Output.MessageOutput.ToolName
}
