package model

// ClientDailyRecord 客户端日统计表
type ClientDailyRecord struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdateAt  int64  `gorm:"autoUpdateTime:milli"`
	Date      string `gorm:"size:16;uniqueIndex:idx_client_daily_date_unique,priority:1"`
	DauCount  int32  `gorm:"index:idx_client_daily_dau_count"`
}
