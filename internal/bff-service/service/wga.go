package service

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetGeneralAgentAssistantSelect(ctx *gin.Context, userId, orgId string, name string) ([]response.GetGeneralAgentAssistantSelectResp, error) {
	return nil, nil
}

func GetGeneralAgentToolSelect(ctx *gin.Context, userId, orgId string) ([]response.GetGeneralAgentToolSelectResp, error) {
	return nil, nil
}

func GetGeneralAgentToolInfo(ctx *gin.Context, userId, orgId string, toolId, toolType string) (*response.GeneralAgentToolInfoResp, error) {
	return nil, nil
}

func CreateGeneralAgentConversation(ctx *gin.Context, userId, orgId string, req request.CreateGeneralAgentConversationReq) (*response.CreateGeneralAgentConversationResp, error) {
	return &response.CreateGeneralAgentConversationResp{ThreadID: uuid.NewString()}, nil
}

func GetGeneralAgentConversationList(ctx *gin.Context, userId, orgId string, req request.GetGeneralAgentConversationListReq) ([]response.GeneralAgentConversationInfo, error) {
	return nil, nil
}

func GetGeneralAgentConversationDetail(ctx *gin.Context, userId, orgId string, req request.GetGeneralAgentConversationDetailReq) (*response.GeneralAgentConversationDetailInfo, error) {
	return nil, nil
}

func DeleteGeneralAgentConversation(ctx *gin.Context, userId, orgId string, req request.DeleteGeneralAgentConversationReq) error {
	return nil
}

func GetGeneralAgentConfig(ctx *gin.Context, userId, orgId string, req request.GetGeneralAgentConfigReq) (*response.GetGeneralAgentConfigResp, error) {
	return nil, nil
}

func CheckGeneralAgentConfig(ctx *gin.Context, userId, orgId string, req request.GeneralAgentConfigCheckRequest) (*response.GeneralAgentConfigCheckResponse, error) {
	return nil, nil
}

func UpdateGeneralAgentConfig(ctx *gin.Context, userId, orgId string, req request.UpdateGeneralAgentConfigReq) error {
	return nil
}

func GeneralAgentWorkspaceDownload(ctx *gin.Context, userId, orgId string, req request.GeneralAgentWorkspaceDownloadReq) (string, []byte, error) {
	return "", nil, nil
}

func GeneralAgentWorkspacePreview(ctx *gin.Context, userId, orgId string, req request.GeneralAgentWorkspacePreviewReq) (string, []byte, string, error) {
	return "", nil, "application/octet-stream", nil
}

func GeneralAgentWorkspaceInfo(ctx *gin.Context, userId, orgId string, req request.GeneralAgentWorkspaceReq) (*response.GeneralAgentWorkspaceResp, error) {
	return nil, nil
}
