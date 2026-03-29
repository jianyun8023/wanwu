package ag_ui_util

// ============================================================================
// 常量定义
// ============================================================================

const (
	RoleAssistant = "assistant"
	RoleUser      = "user"
	RoleTool      = "tool"
	RoleReasoning = "reasoning"
	RoleSystem    = "system"
)

const (
	ToolCallTypeFunction = "function"
)

const (
	ActivityTypeSubAgent  = "sub_agent"
	ActivityTypeWorkspace = "workspace"
)

const (
	ActivityStatusStarted  = "started"
	ActivityStatusFinished = "finished"
)

// ============================================================================
// 消息类型
// ============================================================================

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

type ToolMessage struct {
	MessageID  string `json:"messageId"`
	Role       string `json:"role"`
	ToolCallID string `json:"toolCallId"`
	Content    string `json:"content"`
}

// ============================================================================
// 工具调用类型
// ============================================================================

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ============================================================================
// 活动类型
// ============================================================================

type Activity struct {
	ActivityID   string                 `json:"activityId"`
	ActivityType string                 `json:"activityType"`
	AgentName    string                 `json:"agentName"`
	InstanceNum  int                    `json:"instanceNum"`
	Status       string                 `json:"status"`
	Content      map[string]interface{} `json:"content,omitempty"`
}

// ============================================================================
// 活动内容类型
// ============================================================================

type WorkspaceActivityContent struct {
	RunID     string `json:"runId"`
	ThreadID  string `json:"threadId"`
	FileCount int    `json:"fileCount"`
	TotalSize int64  `json:"totalSize"`
	Timestamp int64  `json:"timestamp"`
}

// ============================================================================
// 工具结果格式化类型
// ============================================================================

type WebSearchResult struct {
	Query    string    `json:"query"`
	WebCount int       `json:"webCount"`
	WebPages []WebPage `json:"webPages"`
}

type WebPage struct {
	Title    string `json:"title"`
	SiteName string `json:"siteName"`
	Icon     string `json:"icon"`
	Summary  string `json:"summary"`
	URL      string `json:"url"`
}
