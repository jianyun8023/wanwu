package model

import (
	"time"
)

type WgaConversation struct {
	ID               uint32    `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:自增Id"`
	ThreadId         string    `gorm:"column:thread_id;uniqueIndex;type:varchar(36);not null;comment:对话ID"`
	Title            string    `gorm:"column:title;type:text;comment:'对话标题'"`
	ConversationType string    `gorm:"column:conversation_type;index:idx_wga_conversation_conversation_type;type:varchar(64);comment:对话类型"`
	UserId           string    `gorm:"column:user_id;index:idx_wga_conversation_user_id;comment:用户id"`
	OrgId            string    `gorm:"column:org_id;index:idx_wga_conversation_org_id;comment:组织id"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime:milli;comment:创建时间"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime:milli;comment:更新时间"`
}

func (WgaConversation) TableName() string {
	return "wga_conversation"
}
