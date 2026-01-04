package models

import (
	"time"

	"gorm.io/gorm"
)

type LedgerEntry struct {
	gorm.Model
	UserID uint `gorm:"not null;index" json:"user_id"` // 外键
	Type   int  `gorm:"type:tinyint;not null;comment:1收入 2支出" json:"type"`
	// 重点：使用 decimal 类型存储金额
	Amount   float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Category string    `gorm:"type:varchar(50);not null" json:"category"`
	Date     time.Time `gorm:"type:date;not null" json:"date"`
	Remark   string    `gorm:"type:varchar(255)" json:"remark"`

	// 建立关联，方便查询
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// EntryFilter 定义了支持的筛选参数
// `form` tag 对应 URL 中的 ?key=value
type EntryFilter struct {
	Type      int    `form:"type"`       // 1收入, 2支出 (0表示全部)
	Category  string `form:"category"`   // 分类名称
	StartDate string `form:"start_date"` // 开始日期 YYYY-MM-DD
	EndDate   string `form:"end_date"`   // 结束日期 YYYY-MM-DD
}
