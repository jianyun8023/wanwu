package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type CreateGeneralAgentConversationResp struct {
	ConversationID string `json:"conversationId"` // 对话ID
}

type GeneralAgentConversationInfo struct {
	ConversationID string `json:"conversationId"` // 对话ID
	Title          string `json:"title"`          // 对话标题
	CreatedAt      string `json:"createdAt"`      // 创建时间
}

type GetGeneralAgentAssistantSelectResp struct {
	AppBriefInfo
}

type GetGeneralAgentToolSelectResp struct {
	Category  string     `json:"category"`  // 类型
	Condition string     `json:"condition"` // 条件 none | optional | required
	ToolList  []ToolInfo `json:"toolList"`  // 工具列表
}

type GeneralAgentToolInfoResp struct {
	ToolName string `json:"toolName"` // 工具名称
	ToolDesc string `json:"toolDesc"` // 工具描述
}

type GetGeneralAgentConfigResp struct {
	ConversationID string                 `json:"conversationId"` // 对话ID
	ModelConfig    request.AppModelConfig `json:"modelConfig"`    // 模型
	AssistantList  []*AssistantAgentInfo  `json:"assistantList"`  // 能体列表
	ToolList       []*AssistantToolInfo   `json:"toolList"`       // 工具列表
}

type GeneralAgentConfigCheckResponse struct {
	Valid     bool             `json:"valid"`     // 是否有效
	ModelMeet bool             `json:"modelMeet"` // 是否符合模型要求
	ToolsMeet []ToolCategories `json:"toolsMeet"` // 工具是否符合要求
}

type ToolCategories struct {
	Category  string      `json:"category"`  // 工具类别类型
	Condition string      `json:"condition"` // 工具类别条件
	Meet      bool        `json:"meet"`      // 是否满足条件
	Tools     []CheckTool `json:"tools"`     // 工具检查结果
}
type CheckTool struct {
	ToolID string `json:"toolId"` // 工具ID
	Meet   bool   `json:"meet"`   // 是否符合要求
}

type FileInfo struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"` // "file" or "directory"
	Size     int64      `json:"size,omitempty"`
	MimeType string     `json:"mimeType,omitempty"`
	Children []FileInfo `json:"children,omitempty"`
}

type GeneralAgentWorkspaceResp struct {
	GeneralAgentConversationWorkspaceInfo
	Path  string     `json:"path"`
	Files []FileInfo `json:"files"`
}

type GeneralAgentConversationDetailInfo struct {
	ConversationID string                 `json:"conversationId"`
	RunID          string                 `json:"runId"`
	CreatedAt      int64                  `json:"createdAt"`
	Messages       []interface{}          `json:"messages"`
	RequestFiles   []AssistantRequestFile `json:"requestFiles"`

	Workspace GeneralAgentConversationWorkspaceInfo `json:"workspace"`
}

type GeneralAgentConversationWorkspaceInfo struct {
	ConversationID string `json:"conversationId"`
	RunID          string `json:"runId"`
	FileCount      int32  `json:"fileCount"`
	TotalSize      int64  `json:"totalSize"`
	IsDisplay      bool   `json:"isDisplay"`
}
