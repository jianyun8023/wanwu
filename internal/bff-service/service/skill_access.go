package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

// canAccessPublishedSkill 判断用户是否有权限访问已发布的 Skill。
// - public: 对所有人可见
// - organization: 对同组织用户可见
// - private: 仅对创建者可见
func canAccessPublishedSkill(appInfo *app_service.AppInfo, userID, orgID string) bool {
	if appInfo == nil {
		return false
	}
	switch appInfo.GetPublishType() {
	case constant.AppPublishPublic:
		return true
	case constant.AppPublishOrganization:
		return appInfo.GetOrgId() == orgID
	case constant.AppPublishPrivate:
		return appInfo.GetUserId() == userID
	default:
		return false
	}
}

// selectAccessiblePublishedSkillAppInfo 从 AppInfo 列表中选择第一条用户可访问的记录。
// 同一 appId 可能有多条发布记录（不同组织发布），按顺序返回第一条可访问的。
func selectAccessiblePublishedSkillAppInfo(appInfos []*app_service.AppInfo, userID, orgID string) *app_service.AppInfo {
	for _, appInfo := range appInfos {
		if canAccessPublishedSkill(appInfo, userID, orgID) {
			return appInfo
		}
	}
	return nil
}

// hasAvailableAcquiredSkillPackage 判断 acquired skill 是否有可用的发布包。
// version 和 objectPath 必须同时非空。
func hasAvailableAcquiredSkillPackage(skill *mcp_service.AcquiredSkill) bool {
	if skill == nil {
		return false
	}
	publish := skill.GetSkill()
	if publish == nil {
		return false
	}
	return publish.GetVersion() != "" && publish.GetObjectPath() != ""
}

// getSourceSkillAppInfoMap 批量获取源 Skill 的发布信息。
// 返回 sourceSkillId 到 AppInfo 列表的映射（同一 Skill 可能有多个发布记录）。
func getSourceSkillAppInfoMap(ctx *gin.Context, sourceSkillIds []string) (map[string][]*app_service.AppInfo, error) {
	if len(sourceSkillIds) == 0 {
		return make(map[string][]*app_service.AppInfo), nil
	}
	appResp, err := app.GetAppListByIds(ctx.Request.Context(), &app_service.GetAppListByIdsReq{
		AppIdsList: sourceSkillIds,
		AppType:    constant.AppTypeSkill,
	})
	if err != nil {
		return nil, err
	}
	result := make(map[string][]*app_service.AppInfo)
	if appResp == nil {
		return result, nil
	}
	for _, info := range appResp.GetInfos() {
		if info == nil || info.GetAppId() == "" {
			continue
		}
		result[info.GetAppId()] = append(result[info.GetAppId()], info)
	}
	return result, nil
}

// isAcquiredSkillAccessible 判断 acquired skill 是否可使用。
// 同时检查：(1) 源 Skill 的发布范围对添加者可见；(2) Skill 有有效的发布包。
// 访问主体为 acquired 记录的所有者（skill.UserId / skill.OrgId）。
func isAcquiredSkillAccessible(ctx *gin.Context, skill *mcp_service.AcquiredSkill, appInfoMap map[string][]*app_service.AppInfo) bool {
	if skill == nil {
		return false
	}
	// 必须有有效的发布包
	if !hasAvailableAcquiredSkillPackage(skill) {
		return false
	}
	// 获取源 Skill ID
	sourceSkillId := skill.GetSkill().GetSkill().GetSkillId()
	if sourceSkillId == "" {
		return false
	}
	// 获取源 Skill 的发布信息列表
	appInfos := appInfoMap[sourceSkillId]
	if len(appInfos) == 0 {
		return false
	}
	// 使用 acquired 记录所有者作为访问主体
	userID := skill.GetUserId()
	orgID := skill.GetOrgId()
	// 判断是否有任一发布记录可访问
	return selectAccessiblePublishedSkillAppInfo(appInfos, userID, orgID) != nil
}

// ensureAcquiredSourceSkillScopeAccessible 校验源 Skill 的发布范围对添加者是否可访问。
// 不可访问时返回带 i18n 的错误。访问主体为 acquired 记录的所有者。
func ensureAcquiredSourceSkillScopeAccessible(ctx *gin.Context, skill *mcp_service.AcquiredSkill) error {
	if skill == nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_not_found", "acquired skill not found")
	}
	sourceSkillId := skill.GetSkill().GetSkill().GetSkillId()
	if sourceSkillId == "" {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_source_id_empty", "acquired skill source custom skill id is empty")
	}
	// 获取源 Skill 的发布信息
	appInfoMap, err := getSourceSkillAppInfoMap(ctx, []string{sourceSkillId})
	if err != nil {
		return err
	}
	appInfos := appInfoMap[sourceSkillId]
	if len(appInfos) == 0 {
		log.Warnf("ensureAcquiredSourceSkillScopeAccessible: source skill not found in app service, sourceSkillId=%s", sourceSkillId)
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_source_inaccessible", "source skill not found in app service")
	}
	// 使用 acquired 记录所有者作为访问主体
	userID := skill.GetUserId()
	orgID := skill.GetOrgId()
	appInfo := selectAccessiblePublishedSkillAppInfo(appInfos, userID, orgID)
	if appInfo == nil {
		log.Warnf("ensureAcquiredSourceSkillScopeAccessible: no accessible appInfo, sourceSkillId=%s, reqUserId=%s, reqOrgId=%s", sourceSkillId, userID, orgID)
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_no_permission", "no permission to access this skill")
	}
	return nil
}
