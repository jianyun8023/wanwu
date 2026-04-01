package orm

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
)

func (c *Client) CreateWgaConversation(ctx context.Context, conversation *model.WgaConversation) *err_code.Status {
	if err := c.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return toErrStatus("wga_conversation_create", err.Error())
	}
	return nil
}

func (c *Client) DeleteWgaConversation(ctx context.Context, threadId string) *err_code.Status {
	if err := sqlopt.WithThreadId(threadId).Apply(c.db.WithContext(ctx)).Delete(&model.WgaConversation{}).Error; err != nil {
		return toErrStatus("wga_conversation_delete", err.Error())
	}
	return nil
}

func (c *Client) GetWgaConversationList(ctx context.Context, conversationType, userID, orgID string, offset, limit int32) ([]*model.WgaConversation, int64, *err_code.Status) {
	var conversations []*model.WgaConversation
	var count int64

	if err := sqlopt.SQLOptions(
		sqlopt.WithUserId(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithConversationType(conversationType),
	).Apply(c.db.WithContext(ctx).Model(&model.WgaConversation{})).Offset(int(offset)).Limit(int(limit)).Order("created_at DESC").Find(&conversations).Error; err != nil {
		return conversations, count, toErrStatus("wga_conversation_list", err.Error())
	}

	return conversations, int64(len(conversations)), nil
}

func (c *Client) WgaConversationExists(ctx context.Context, threadId, userID, orgID string) (bool, *err_code.Status) {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserId(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithThreadId(threadId),
	).Apply(c.db.WithContext(ctx).Model(&model.WgaConversation{})).Count(&count).Error; err != nil {
		return false, toErrStatus("wga_conversation_exists", err.Error())
	}
	return count > 0, nil
}
