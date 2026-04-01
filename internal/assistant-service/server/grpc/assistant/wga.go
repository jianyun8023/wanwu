package assistant

import (
	"context"
	"encoding/json"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetWgaConfig(ctx context.Context, req *assistant_service.GetWgaConfigReq) (*assistant_service.GetWgaConfigResp, error) {
	if req.ThreadId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigGetErr, "thread_id is required")
	}
	if req.Identity == nil {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigGetErr, "identity is required")
	}

	config, status := s.cli.GetWgaConfig(ctx, req.ThreadId, req.Identity.UserId, req.Identity.OrgId)
	if status != nil {
		log.Errorf("获取WGA配置失败，conversationId: %s, error: %v", req.ThreadId, status)
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	return &assistant_service.GetWgaConfigResp{
		Config: toProtoWgaConfig(config),
	}, nil
}

func (s *Service) UpdateWgaConfig(ctx context.Context, req *assistant_service.UpdateWgaConfigReq) (*emptypb.Empty, error) {
	if req.ThreadId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigUpdateErr, "thread_id is required")
	}
	if req.Identity == nil {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigUpdateErr, "identity is required")
	}

	toolListJSON, _ := json.Marshal(req.ToolList)
	assistantListJSON, _ := json.Marshal(req.AssistantList)

	var modelConfigJSON string
	if req.ModelConfig != nil {
		modelConfigBytes, _ := json.Marshal(req.ModelConfig)
		modelConfigJSON = string(modelConfigBytes)
	}

	config := &model.WgaConfig{
		ThreadID:      req.ThreadId,
		UserID:        req.Identity.UserId,
		OrgID:         req.Identity.OrgId,
		ModelConfig:   modelConfigJSON,
		AssistantList: string(assistantListJSON),
		ToolList:      string(toolListJSON),
	}

	status := s.cli.UpdateWgaConfig(ctx, config)
	if status != nil {
		log.Errorf("更新WGA配置失败，conversationId: %s, error: %v", req.ThreadId, status)
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	return &emptypb.Empty{}, nil
}

func toProtoWgaConfig(m *model.WgaConfig) *assistant_service.WgaConfig {
	config := &assistant_service.WgaConfig{
		ThreadId:  m.ThreadID,
		UserId:    m.UserID,
		OrgId:     m.OrgID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.ModelConfig != "" {
		var modelConfig assistant_service.WgaModelConfig
		if err := json.Unmarshal([]byte(m.ModelConfig), &modelConfig); err == nil {
			config.ModelConfig = &modelConfig
		}
	}

	if m.ToolList != "" {
		var tools []assistant_service.WgaConfigTool
		if err := json.Unmarshal([]byte(m.ToolList), &tools); err == nil {
			for _, t := range tools {
				config.ToolList = append(config.ToolList, &assistant_service.WgaConfigTool{
					ToolId:   t.ToolId,
					ToolType: t.ToolType,
				})
			}
		}
	}

	if m.AssistantList != "" {
		var assistants []assistant_service.WgaConfigAssistant
		if err := json.Unmarshal([]byte(m.AssistantList), &assistants); err == nil {
			for _, a := range assistants {
				config.AssistantList = append(config.AssistantList, &assistant_service.WgaConfigAssistant{
					AssistantId:   a.AssistantId,
					AssistantType: a.AssistantType,
				})
			}
		}
	}

	return config
}
