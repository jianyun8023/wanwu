package response

type AgentProxyChatResp struct {
	Code      int                  `json:"code"`
	Message   string               `json:"message"`
	Response  string               `json:"response"`
	Order     int                  `json:"order"`
	EventType int                  `json:"eventType"`
	EventData *AgentProxyEventData `json:"eventData"`
	Finish    int                  `json:"finish"`
	Usage     *AgentProxyUsage     `json:"usage"`
}

type AgentProxyEventData struct {
	Status    int    `json:"status"`
	Id        string `json:"id"`
	EventType int    `json:"eventType"`
	Name      string `json:"name"`
	Profile   string `json:"profile"`
	TimeCost  string `json:"timeCost"`
	ParentId  string `json:"parentId"`
	Order     int    `json:"order"`
}

type AgentProxyUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
