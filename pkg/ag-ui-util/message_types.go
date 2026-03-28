package ag_ui_util

const (
	RoleAssistant = "assistant"
	RoleUser      = "user"
	RoleTool      = "tool"
	RoleReasoning = "reasoning"
	RoleSystem    = "system"

	ToolCallTypeFunction = "function"
	ActivityTypeSubAgent = "sub_agent"

	ActivityStatusStarted  = "started"
	ActivityStatusFinished = "finished"
)

type TextMessage struct {
	MessageID string `json:"messageId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

type ReasoningMessage struct {
	MessageID string `json:"messageId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolMessage struct {
	MessageID  string `json:"messageId"`
	Role       string `json:"role"`
	ToolCallID string `json:"toolCallId"`
	Content    string `json:"content"`
}

type Activity struct {
	ActivityID   string                 `json:"activityId"`
	ActivityType string                 `json:"activityType"`
	AgentName    string                 `json:"agentName"`
	InstanceNum  int                    `json:"instanceNum"`
	Status       string                 `json:"status"`
	Content      map[string]interface{} `json:"content,omitempty"`
}
