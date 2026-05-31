// Package model 定义 GORM 数据模型与行情值对象。
// 金额/价格统一使用 shopspring/decimal，避免浮点误差（见 BACKEND.md §3.4）。
package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 公共基类：id / created_at / updated_at / deleted_at（软删除）。
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
