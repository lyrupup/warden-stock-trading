package response

import (
	"time"

	"github.com/shopspring/decimal"

	"warden/internal/strategy/rule"
)

// Strategy 策略详情/列表项（对齐 openapi Strategy）。
type Strategy struct {
	ID           uint            `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Tags         string          `json:"tags"`
	Status       int8            `json:"status"`
	Indicators   *rule.RuleGroup `json:"indicators,omitempty"`
	Skill        string          `json:"skill,omitempty"`
	SkillVersion int             `json:"skill_version,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// ScreenCandidate 粗筛命中候选（数值字段为 decimal 字符串）。
type ScreenCandidate struct {
	StockCode string                     `json:"stock_code"`
	StockName string                     `json:"stock_name"`
	Score     decimal.Decimal            `json:"score"`
	Factors   map[string]decimal.Decimal `json:"factors"`
	Matched   []int                      `json:"matched"`
}

// ScreenResult 粗筛结果（对齐 openapi ScreenResult）。
type ScreenResult struct {
	ID            uint              `json:"id"`
	StrategyID    uint              `json:"strategy_id"`
	Status        int8              `json:"status"`
	TradeDate     string            `json:"trade_date,omitempty"`
	UniverseCount int               `json:"universe_count"`
	MatchedCount  int               `json:"matched_count"`
	Candidates    []ScreenCandidate `json:"candidates"`
	ErrorMsg      string            `json:"error_msg,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}
