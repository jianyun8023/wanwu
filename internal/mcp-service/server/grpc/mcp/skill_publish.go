package mcp

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) CreatePublishCustomSkill(ctx context.Context, req *mcp_service.PublishCustomSkillReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, toErrStatus("mcp_custom_skill_publish", "identity is empty"))
	}
	var snap *orm.CustomSkillPublishSnapshot
	if req.GetMarkdown() != "" || req.GetObjectPath() != "" {
		snap = &orm.CustomSkillPublishSnapshot{
			Markdown:   req.GetMarkdown(),
			ObjectPath: req.GetObjectPath(),
		}
	}
	publish := &model.CustomSkillPublish{
		SkillID:            req.GetSkillId(),
		Version:            req.GetVersion(),
		VersionDescription: req.GetVersionDesc(),
	}
	if err := s.cli.PublishCustomSkill(ctx, publish, snap); err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdatePublishCustomSkill(ctx context.Context, req *mcp_service.UpdatePublishCustomSkillReq) (*emptypb.Empty, error) {
	if err := s.cli.UpdatePublishCustomSkill(ctx, req.SkillId, req.VersionDesc); err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetPublishCustomSkillHistoryList(ctx context.Context, req *mcp_service.GetPublishCustomSkillHistoryListReq) (*mcp_service.PublishCustomSkillHistoryListResp, error) {
	history, total, err := s.cli.GetPublishCustomSkillHistoryList(ctx, req.SkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	customSkill, csErr := s.cli.GetCustomSkill(ctx, req.SkillId)
	if csErr != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, csErr)
	}
	list := make([]*mcp_service.PublishCustomSkill, 0, len(history))
	for _, item := range history {
		list = append(list, toProtoPublishCustomSkill(customSkill, item))
	}
	return &mcp_service.PublishCustomSkillHistoryListResp{HistoryList: list, Total: total}, nil
}

func (s *Service) GetPublishCustomSkillByLatest(ctx context.Context, req *mcp_service.GetPublishCustomSkillByLatestReq) (*mcp_service.PublishCustomSkill, error) {
	customSkill, csErr := s.cli.GetCustomSkill(ctx, req.SkillId)
	if csErr != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, csErr)
	}
	publish, err := s.cli.GetPublishCustomSkillByLatest(ctx, req.SkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return toProtoPublishCustomSkill(customSkill, publish), nil
}

func (s *Service) GetPublishCustomSkillByVersion(ctx context.Context, req *mcp_service.GetPublishCustomSkillByVersionReq) (*mcp_service.PublishCustomSkill, error) {
	customSkill, csErr := s.cli.GetCustomSkill(ctx, req.SkillId)
	if csErr != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, csErr)
	}
	publish, err := s.cli.GetPublishCustomSkillByVersion(ctx, req.SkillId, req.Version)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return toProtoPublishCustomSkill(customSkill, publish), nil
}

// GetPublishCustomSkillList 与 CustomSkillGetList 共用 buildCustomSkillGetListResp（草稿 + 可选最新发布）。
func (s *Service) GetPublishCustomSkillList(ctx context.Context, req *mcp_service.GetPublishCustomSkillListReq) (*mcp_service.GetPublishCustomSkillListResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, toErrStatus("mcp_custom_skill_publish_list", "identity is empty"))
	}
	resp, st := s.listPublishCustomSkills(ctx, req.Identity.UserId, req.Identity.OrgId, req.Name)
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	return &mcp_service.GetPublishCustomSkillListResp{List: resp.List}, nil
}

func (s *Service) GetPublishCustomSkillByIDList(ctx context.Context, req *mcp_service.GetPublishCustomSkillByIDListReq) (*mcp_service.GetPublishCustomSkillByIDListResp, error) {
	customSkills, err := s.cli.GetCustomSkillBySkillIds(ctx, req.SkillIdList)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	resp, st := s.buildCustomSkillGetListResp(ctx, customSkills)
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	return &mcp_service.GetPublishCustomSkillByIDListResp{List: resp.List}, nil
}
