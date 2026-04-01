package orm

import (
	"context"
	"errors"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) GetWgaConfig(ctx context.Context, threadId string, userId, orgId string) (*model.WgaConfig, *err_code.Status) {
	var config model.WgaConfig
	if err := sqlopt.SQLOptions(
		sqlopt.WithThreadId(threadId),
		sqlopt.WithUserId(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Model(&model.WgaConfig{}).First(&config).Error; err != nil {
		return nil, toErrStatus("wga_config_get", err.Error())
	}
	return &config, nil
}

func (c *Client) UpdateWgaConfig(ctx context.Context, config *model.WgaConfig) *err_code.Status {
	if err := sqlopt.WithThreadId(config.ThreadID).Apply(c.db.WithContext(ctx)).First(&model.WgaConversation{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("wga_config_update", "conversation not found")
		}
		return toErrStatus("wga_config_check_conversation", err.Error())
	}

	var existing model.WgaConfig
	err := sqlopt.SQLOptions(
		sqlopt.WithThreadId(config.ThreadID),
		sqlopt.WithUserId(config.UserID),
		sqlopt.WithOrgID(config.OrgID),
	).Apply(c.db.WithContext(ctx)).First(&existing).Error

	if err == nil {
		result := c.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"thread_id":      config.ThreadID,
			"model_config":   config.ModelConfig,
			"assistant_list": config.AssistantList,
			"tool_list":      config.ToolList,
		})
		if result.Error != nil {
			return toErrStatus("wga_config_update", result.Error.Error())
		}
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return toErrStatus("wga_config_get", err.Error())
	}

	if err := c.db.WithContext(ctx).Create(config).Error; err != nil {
		return toErrStatus("wga_config_create", err.Error())
	}
	return nil
}
