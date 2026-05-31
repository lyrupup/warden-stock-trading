package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// WatchlistItem 自选股（M1）。所有查询强制带 user_id，杜绝越权。
type WatchlistItem struct {
	BaseModel
	UserID    uint   `gorm:"index;not null" json:"user_id"`
	StockCode string `gorm:"size:16;not null" json:"stock_code"`
	StockName string `gorm:"size:64;not null;default:''" json:"stock_name"`
	GroupName string `gorm:"size:64;not null;default:'default'" json:"group_name"`
	Remark    string `gorm:"size:256;not null;default:''" json:"remark"`
	Sort      int    `gorm:"not null;default:0" json:"sort"`
}

func (WatchlistItem) TableName() string { return "watchlist_items" }

// IndexQuote 指数行情快照（公共，不归属单一用户，缓存兜底）。
type IndexQuote struct {
	ID            uint            `gorm:"primarykey" json:"id"`
	IndexCode     string          `gorm:"size:16;not null" json:"index_code"`
	IndexName     string          `gorm:"size:64;not null" json:"index_name"`
	Price         decimal.Decimal `gorm:"type:numeric(20,4)" json:"price"`
	ChangeAmount  decimal.Decimal `gorm:"type:numeric(20,4)" json:"change_amount"`
	ChangePercent decimal.Decimal `gorm:"type:numeric(10,4)" json:"change_percent"`
	Volume        decimal.Decimal `gorm:"type:numeric(20,0)" json:"volume"`
	Amount        decimal.Decimal `gorm:"type:numeric(24,4)" json:"amount"`
	TradeDate     time.Time       `gorm:"type:date;not null" json:"trade_date"`
	SnapshotAt    time.Time       `gorm:"not null;default:now()" json:"snapshot_at"`
}

func (IndexQuote) TableName() string { return "index_quotes" }

// StockQuote 个股行情快照（公共，缓存兜底）。
// Stale 字段为降级标记（不持久化）：外部源失败返回最近快照时置 true。
type StockQuote struct {
	ID            uint            `gorm:"primarykey" json:"id"`
	StockCode     string          `gorm:"size:16;not null" json:"stock_code"`
	StockName     string          `gorm:"size:64;not null;default:''" json:"stock_name"`
	Price         decimal.Decimal `gorm:"type:numeric(20,4)" json:"price"`
	Open          decimal.Decimal `gorm:"type:numeric(20,4)" json:"open"`
	High          decimal.Decimal `gorm:"type:numeric(20,4)" json:"high"`
	Low           decimal.Decimal `gorm:"type:numeric(20,4)" json:"low"`
	PrevClose     decimal.Decimal `gorm:"type:numeric(20,4)" json:"prev_close"`
	ChangePercent decimal.Decimal `gorm:"type:numeric(10,4)" json:"change_percent"`
	Volume        decimal.Decimal `gorm:"type:numeric(20,0)" json:"volume"`
	Amount        decimal.Decimal `gorm:"type:numeric(24,4)" json:"amount"`
	TurnoverRate  decimal.Decimal `gorm:"type:numeric(10,4)" json:"turnover_rate"`
	TradeDate     time.Time       `gorm:"type:date;not null" json:"trade_date"`
	SnapshotAt    time.Time       `gorm:"not null;default:now()" json:"snapshot_at"`
	Stale         bool            `gorm:"-" json:"stale"`
}

func (StockQuote) TableName() string { return "stock_quotes" }

// Kline 是 K 线值对象（不落 GORM 表，由行情源实时计算/返回）。
type Kline struct {
	Date   string          `json:"date"`
	Open   decimal.Decimal `json:"open"`
	High   decimal.Decimal `json:"high"`
	Low    decimal.Decimal `json:"low"`
	Close  decimal.Decimal `json:"close"`
	Volume decimal.Decimal `json:"volume"`
	Amount decimal.Decimal `json:"amount"`
}

// StockBrief 是股票搜索结果的精简值对象。
type StockBrief struct {
	StockCode string `json:"stock_code"`
	StockName string `json:"stock_name"`
	Market    string `json:"market"`
}
