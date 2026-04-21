package opencode

import "encoding/json"

// ============================================================================
// Opencode 输出类型（导出）
// ============================================================================

// OpencodeEventType 事件类型。
type OpencodeEventType string

// 事件类型常量。
const (
	OpencodeEventTypeStepStart  OpencodeEventType = "step_start"
	OpencodeEventTypeStepFinish OpencodeEventType = "step_finish"
	OpencodeEventTypeText       OpencodeEventType = "text"
	OpencodeEventTypeToolUse    OpencodeEventType = "tool_use"
	OpencodeEventTypeReasoning  OpencodeEventType = "reasoning"
	OpencodeEventTypeSnapshot   OpencodeEventType = "snapshot"
	OpencodeEventTypePatch      OpencodeEventType = "patch"
	OpencodeEventTypeFile       OpencodeEventType = "file"
	OpencodeEventTypeAgent      OpencodeEventType = "agent"
	OpencodeEventTypeRetry      OpencodeEventType = "retry"
	OpencodeEventTypeSubtask    OpencodeEventType = "subtask"
	OpencodeEventTypeCompaction OpencodeEventType = "compaction"
	OpencodeEventTypeError      OpencodeEventType = "error"
)

// OpencodeEvent 输出事件结构。
type OpencodeEvent struct {
	Type      OpencodeEventType `json:"type"`
	Timestamp int64             `json:"timestamp"`
	SessionID string            `json:"sessionID"`
	Part      json.RawMessage   `json:"part"`
}

// ============================================================================
// 事件部分类型
// ============================================================================

// textPart 文本部分。
type textPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// reasoningPart 推理部分。
type reasoningPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// toolPart 工具调用部分。
type toolPart struct {
	Type   string    `json:"type"`
	CallID string    `json:"callID,omitempty"`
	Tool   string    `json:"tool"`
	State  toolState `json:"state"`
}

// toolState 工具调用状态。
type toolState struct {
	Status string      `json:"status"`
	Input  interface{} `json:"input,omitempty"`
	Output string      `json:"output,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// stepStartPart 步骤开始部分。
type stepStartPart struct {
	Type string `json:"type"`
}

// stepFinishPart 步骤结束部分。
type stepFinishPart struct {
	Type   string               `json:"type"`
	Reason string               `json:"reason,omitempty"`
	Tokens stepFinishPartTokens `json:"tokens,omitempty"`
	Cost   float64              `json:"cost,omitempty"`
}

// stepFinishPartTokens 步骤结束 token 统计。
type stepFinishPartTokens struct {
	Input     float64 `json:"input,omitempty"`
	Output    float64 `json:"output,omitempty"`
	Reasoning float64 `json:"reasoning,omitempty"`
	Cache     struct {
		Read  float64 `json:"read,omitempty"`
		Write float64 `json:"write,omitempty"`
	} `json:"cache,omitempty"`
}

// errorPart 错误部分。
type errorPart struct {
	Error struct {
		Name string `json:"name"`
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	} `json:"error"`
}

// ErrorPart 错误部分（导出）。
type ErrorPart = errorPart

// ============================================================================
// SSE 事件类型（内部使用）
// ============================================================================

// sseEvent SSE 事件结构。
type sseEvent struct {
	Directory string          `json:"directory"`
	Payload   sseEventPayload `json:"payload"`
}

// sseEventPayload SSE 事件负载。
type sseEventPayload struct {
	Type       string        `json:"type"`
	Properties sseEventProps `json:"properties"`
}

// sseEventProps SSE 事件属性。
type sseEventProps struct {
	SessionID string         `json:"sessionID"`
	Delta     string         `json:"delta"` // 旧版 message.part.updated 的 delta（1.3.17 已废弃）
	Part      sseEventPart   `json:"part"`
	Status    sseEventStatus `json:"status"`
	Error     sseEventError  `json:"error"`
	Info      sseEventInfo   `json:"info"`
	// 1.3.17 新增：message.part.delta 事件字段
	MessageID string `json:"messageID"`
	PartID    string `json:"partID"`
	Field     string `json:"field"`
}

// sseEventInfo SSE 消息信息（message.updated 事件）。
type sseEventInfo struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// sseEventError SSE 错误事件属性。
type sseEventError struct {
	Name string `json:"name"`
	Data struct {
		Message    string `json:"message"`
		StatusCode int    `json:"statusCode,omitempty"`
	} `json:"data"`
}

// sseEventPart SSE 事件部分。
type sseEventPart struct {
	ID        string       `json:"id"`
	SessionID string       `json:"sessionID"`
	MessageID string       `json:"messageID"`
	Type      string       `json:"type"`
	Text      string       `json:"text"`
	CallID    string       `json:"callID"`
	Tool      string       `json:"tool"`
	State     sseToolState `json:"state"`
	Reason    string       `json:"reason"` // step-finish 的 reason
	Tokens    struct {
		Total     float64 `json:"total"`
		Input     float64 `json:"input"`
		Output    float64 `json:"output"`
		Reasoning float64 `json:"reasoning"`
		Cache     struct {
			Read  float64 `json:"read"`
			Write float64 `json:"write"`
		} `json:"cache"`
	} `json:"tokens"` // step-finish 的 token 统计
	Cost float64 `json:"cost"` // step-finish 的 cost
}

// sseToolState 工具调用状态。
type sseToolState struct {
	Status string      `json:"status"`
	Input  interface{} `json:"input,omitempty"`
	Output string      `json:"output,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// sseEventStatus SSE 事件状态。
type sseEventStatus struct {
	Type string `json:"type"`
}
