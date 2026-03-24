package request

type CreateGeneralAgentConversationReq struct {
	Title string `json:"title" validate:"required"` // 标题
}

func (c *CreateGeneralAgentConversationReq) Check() error { return nil }

type DeleteGeneralAgentConversationReq struct {
	ConversationID string `json:"conversationId" validate:"required"` // 对话ID
}

func (c *DeleteGeneralAgentConversationReq) Check() error { return nil }

type GetGeneralAgentConversationListReq struct {
	PageNo   int `json:"pageNo" form:"pageNo" validate:"required"`     // 页码
	PageSize int `json:"pageSize" form:"pageSize" validate:"required"` // 每页数量
}

func (c *GetGeneralAgentConversationListReq) Check() error { return nil }

type GetGeneralAgentConversationDetailReq struct {
	ConversationID string `json:"conversationId" validate:"required"` // 对话ID
}

func (c *GetGeneralAgentConversationDetailReq) Check() error { return nil }

type GetGeneralAgentConfigReq struct {
	ConversationID string `json:"conversationId" validate:"required"` // 对话ID
}

func (c *GetGeneralAgentConfigReq) Check() error { return nil }

type GeneralAgentConfigCheckRequest struct {
	ConversationID string `json:"conversationId" validate:"required"` // 对话ID
}

func (c *GeneralAgentConfigCheckRequest) Check() error { return nil }

type UpdateGeneralAgentConfigReq struct {
	ConversationID string              `json:"conversationId" validate:"required"` // 对话ID
	ModelConfig    *AppModelConfig     `json:"modelConfig" validate:"required"`    // 模型
	ToolList       []ToolSelected      `json:"toolList"`                           // 工具id
	AssistantList  []AssistantSelected `json:"assistantList"`                      // 智能体id
}

type ToolSelected struct {
	ToolID   string `json:"toolId" validate:"required"`   // 工具id
	ToolType string `json:"toolType" validate:"required"` // 工具类型
}

type AssistantSelected struct {
	AssistantID       string `json:"assistantId" validate:"required"`       // 智能体id
	AssistantCategory string `json:"assistantCategory" validate:"required"` // 智能体类型
}

func (c *UpdateGeneralAgentConfigReq) Check() error { return nil }

type GeneralAgentConversationChatReq struct {
	ConversationID string                 `json:"conversationId" validate:"required"` // 对话ID
	Query          string                 `json:"query" validate:"required"`          // 用户问题
	FileInfo       []ConversionStreamFile `json:"fileInfo" form:"fileInfo"`           // 文件信息
}

func (c *GeneralAgentConversationChatReq) Check() error { return nil }

type GeneralAgentWorkspaceDownloadReq struct {
	ConversationID string `json:"conversationId" validate:"required"` // 对话ID
	RunID          string `json:"runId" validate:"required"`          // 运行id
	Path           string `json:"path"`                               // workspace路径
}

func (c *GeneralAgentWorkspaceDownloadReq) Check() error { return nil }

type GeneralAgentWorkspacePreviewReq struct {
	ConversationID string `json:"conversationId" validate:"required"` // 对话ID
	RunID          string `json:"runId" validate:"required"`          // 运行id
	Path           string `json:"path" validate:"required"`           // 文件路径
}

func (c *GeneralAgentWorkspacePreviewReq) Check() error { return nil }

type GeneralAgentWorkspaceReq struct {
	ConversationID string `json:"conversationId" validate:"required"` // 对话ID
	RunID          string `json:"runId" validate:"required"`          // 运行id
}

func (c *GeneralAgentWorkspaceReq) Check() error { return nil }
