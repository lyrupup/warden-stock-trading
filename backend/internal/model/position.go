package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Position 持仓（M3）。盈亏口径为移动加权平均成本（见 BACKEND.md §M3）。
type Position struct {
	BaseModel
	UserID      uint            `gorm:"index;not null" json:"user_id"`
	StockCode   string          `gorm:"size:16;not null" json:"stock_code"`
	StockName   string          `gorm:"size:64;not null;default:''" json:"stock_name"`
	Quantity    decimal.Decimal `gorm:"type:numeric(20,4)" json:"quantity"`
	AvgCost     decimal.Decimal `gorm:"type:numeric(20,4)" json:"avg_cost"`
	TotalCost   decimal.Decimal `gorm:"type:numeric(20,4)" json:"total_cost"`
	RealizedPnL decimal.Decimal `gorm:"type:numeric(20,4)" json:"realized_pnl"`
	Status      int8            `gorm:"not null;default:1" json:"status"` // 1 持有 2 已清仓
	OpenedAt    time.Time       `json:"opened_at"`
	ClosedAt    *time.Time      `json:"closed_at,omitempty"`
}

func (Position) TableName() string { return "positions" }

// Trade 交易记录（M3）。卖出时计算本笔已实现盈亏。
type Trade struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	PositionID  uint            `gorm:"index;not null" json:"position_id"`
	UserID      uint            `gorm:"index;not null" json:"user_id"`
	Side        int8            `gorm:"not null" json:"side"` // 1 买入 2 卖出
	Price       decimal.Decimal `gorm:"type:numeric(20,4);not null" json:"price"`
	Quantity    decimal.Decimal `gorm:"type:numeric(20,4);not null" json:"quantity"`
	Fee         decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"fee"`
	Tax         decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"tax"`
	RealizedPnL decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"realized_pnl"`
	Remark      string          `gorm:"size:256;not null;default:''" json:"remark"`
	TradedAt    time.Time       `gorm:"not null" json:"traded_at"`
	CreatedAt   time.Time       `json:"created_at"`
}

func (Trade) TableName() string { return "trades" }
