package assistant

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// WgaConversationCreate 创建WGA对话
func (s *Service) WgaConversationCreate(ctx context.Context, req *assistant_service.WgaConversationCreateReq) (*assistant_service.WgaConversationCreateResp, error) {
	// 创建对话
	threadId := util.GenUUID()
	conversation := &model.WgaConversation{
		ThreadId:         threadId,
		Title:            req.Prompt,
		ConversationType: req.ConversationType,
		UserId:           req.Identity.UserId,
		OrgId:            req.Identity.OrgId,
	}

	if status := s.cli.CreateWgaConversation(ctx, conversation); status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	// 创建对话配置
	_, err := s.UpdateWgaConversationConfig(ctx, &assistant_service.UpdateWgaConversationConfigReq{
		ThreadId:    threadId,
		ModelConfig: req.ModelConfig,
		Identity:    req.Identity,
	})
	if err != nil {
		return nil, err
	}

	return &assistant_service.WgaConversationCreateResp{
		ThreadId: threadId,
	}, nil
}

// WgaConversationDelete 删除WGA对话
func (s *Service) WgaConversationDelete(ctx context.Context, req *assistant_service.WgaConversationDeleteReq) (*emptypb.Empty, error) {
	if status := s.cli.DeleteWgaConversation(ctx, req.ThreadId); status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	return &emptypb.Empty{}, nil
}

// WgaConversationList 获取WGA对话列表
func (s *Service) WgaConversationList(ctx context.Context, req *assistant_service.WgaConversationListReq) (*assistant_service.WgaConversationListResp, error) {
	offset := (req.PageNo - 1) * req.PageSize

	conversations, total, status := s.cli.GetWgaConversationList(ctx, req.ConversationType, req.Identity.UserId, req.Identity.OrgId, offset, req.PageSize)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	var conversationInfos []*assistant_service.WgaConversationInfo
	for _, conversation := range conversations {
		conversationInfos = append(conversationInfos, &assistant_service.WgaConversationInfo{
			ThreadId:  conversation.ThreadId,
			Title:     conversation.Title,
			CreatedAt: conversation.CreatedAt,
		})
	}

	return &assistant_service.WgaConversationListResp{
		Data:     conversationInfos,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

// WgaConversationExists 检查WGA对话是否存在
func (s *Service) WgaConversationExists(ctx context.Context, req *assistant_service.WgaConversationExistsReq) (*assistant_service.WgaConversationExistsResp, error) {
	exists, status := s.cli.WgaConversationExists(ctx, req.ThreadId, req.Identity.UserId, req.Identity.OrgId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	return &assistant_service.WgaConversationExistsResp{
		Exists: exists,
	}, nil
}
