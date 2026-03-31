package request

type AgentProxyChatReq struct {
	AssistantId uint32   `json:"assistantId" validate:"required"`
	UserId      string   `json:"userId" validate:"required"`
	OrgId       string   `json:"orgId" validate:"required"`
	Input       string   `json:"input" validate:"required"`
	UploadFile  []string `json:"uploadFile"`
}

func (r *AgentProxyChatReq) Check() error {
	return nil
}
