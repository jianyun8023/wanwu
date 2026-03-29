// Package ag_ui_util 提供 AG-UI 协议的事件转换功能。
//
// 本实现采用串行处理模式，是 AG-UI 完整规范的子集。
// 完整的 AG-UI 协议规范请参考 README.md。
package ag_ui_util

import (
	"context"
	"encoding/json"

	"github.com/UnicomAI/wanwu/pkg/util"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
)

// BaseState 维护 AG-UI 事件流的基础状态。
// 跟踪 Run 和消息的活跃状态，确保事件的正确顺序。
type BaseState struct {
	threadID    string
	runID       string
	runStarted  bool
	runFinished bool
	*MessageState
}

// NewBaseState 创建基础状态。
func NewBaseState(threadID, runID string) BaseState {
	return BaseState{
		threadID:     threadID,
		runID:        runID,
		MessageState: NewMessageState(),
	}
}

func (s *BaseState) ThreadID() string    { return s.threadID }
func (s *BaseState) RunID() string       { return s.runID }
func (s *BaseState) MessageID() string   { return s.TextMsgID() }
func (s *BaseState) ReasoningID() string { return s.MessageState.ReasoningID() }
func (s *BaseState) ReasoningMessageID() string {
	return s.ReasoningMsgID()
}

func (s *BaseState) SetMessageID(messageID string) {
	s.SetTextMsgID(messageID)
}

// ResetMessageID 重置消息状态，用于 Tool 消息处理完毕后准备接收新的 Assistant 消息。
//
// 根据 AG-UI 协议：
//   - TOOL_CALL_RESULT 创建独立的 ToolMessage，不自动重置 TextMessage/Reasoning 状态
//   - 但 Tool 消息意味着当前轮次结束，后续 Assistant 响应应使用新的 messageId
//   - 同时重置 bool 状态是防御性编程，确保即使未调用 End 方法也能正确重置
func (s *BaseState) ResetMessageID() {
	s.Reset()
}

// EnsureRunStarted 确保 RunStarted 事件已发送（幂等）。
// AG-UI 协议要求：Run 开始时必须发送 RUN_STARTED 事件。
func (s *BaseState) EnsureRunStarted() []aguievents.Event {
	if s.runStarted {
		return nil
	}
	s.runStarted = true
	return []aguievents.Event{aguievents.NewRunStartedEvent(s.threadID, s.runID)}
}

// FinishBase 生成基础结束事件，确保所有活跃的消息状态被正确关闭。
// 调用顺序：RUN_STARTED → EndAll → RUN_FINISHED
//
// AG-UI 协议要求：所有活跃的消息必须在进行 RUN_FINISHED 之前正确关闭。
func (s *BaseState) FinishBase() []aguievents.Event {
	if s.runFinished {
		return nil
	}
	s.runFinished = true
	var events []aguievents.Event
	events = append(events, s.EnsureRunStarted()...)
	events = append(events, s.EndAll()...)
	events = append(events, aguievents.NewRunFinishedEvent(s.threadID, s.runID))
	return events
}

// EventsToJSONChannel 将事件流转换为 JSON 字符串流，用于 SSE 等传输场景。
func EventsToJSONChannel(ctx context.Context, in <-chan aguievents.Event) <-chan string {
	out := make(chan string, 1024)
	go func() {
		defer util.PrintPanicStack()
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-in:
				if !ok {
					return
				}
				if data, err := json.Marshal(evt); err == nil {
					out <- string(data)
				}
			}
		}
	}()
	return out
}
