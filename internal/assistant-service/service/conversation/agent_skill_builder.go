package conversation

import (
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
)

var agentSkill = &AgentSkill{}

type AgentSkill struct {
}

func init() {
	InitBuilder(agentSkill)
}

func (*AgentSkill) EventType() int {
	return SkillEventType
}
func (*AgentSkill) Build(conversationResp *ConversationResp, conversation, searchResult string, agentChatResp *AgentChatResp) error {
	eventData := agentChatResp.EventData
	if eventData == nil {
		return nil
	}
	resp := conversationResp.ConversationEventMap[eventData.Id]
	if resp == nil {
		resp = CreateConversationResp()
		resp.Order = eventData.Order
		resp.EventType = SkillEventType
		conversationResp.ConversationEventMap[eventData.Id] = resp
	}
	if resp.SearchList == nil && len(searchResult) > 0 {
		resp.SearchList = &searchResult
	}
	if len(conversation) > 0 {
		//保存对话
		resp.Write(conversation, eventData.Order)
	}
	//终态存储
	if eventData.Status == model.EventEndStatus || eventData.Status == model.EventFailStatus {
		resp.EventData = eventData
	}
	return nil
}
