package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSON 是用于 PostgreSQL jsonb 列的通用类型，承载量化规则组、候选列表等结构化数据。
// 以原始 JSON 存储，避免 model 包反向依赖 strategy/factor/rule（防止 import cycle）；
// 由 service 层负责与 rule.RuleGroup 等业务类型互转。
type JSON json.RawMessage

// Value 实现 driver.Valuer，写入 jsonb。
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan 实现 sql.Scanner，从 jsonb 读取。
func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		b := make([]byte, len(v))
		copy(b, v)
		*j = b
	case string:
		*j = []byte(v)
	default:
		return errors.New("model.JSON: 不支持的 Scan 类型")
	}
	return nil
}

// MarshalJSON 让 JSON 字段在响应里按原始 JSON 输出。
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON 直接保存原始字节。
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("model.JSON: UnmarshalJSON on nil pointer")
	}
	b := make([]byte, len(data))
	copy(b, data)
	*j = b
	return nil
}

// 策略状态。
const (
	StrategyStatusEnabled  int8 = 1
	StrategyStatusDisabled int8 = 0
)

// 粗筛任务状态（复用 backtest 同款状态机）。
const (
	ScreenStatusPending int8 = 0
	ScreenStatusRunning int8 = 1
	ScreenStatusDone    int8 = 2
	ScreenStatusFailed  int8 = 3
)

// Strategy 选股策略（M2）。所有查询强制带 user_id，杜绝越权。
type Strategy struct {
	BaseModel
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	Name        string `gorm:"size:128;not null" json:"name"`
	Description string `gorm:"type:text;not null;default:''" json:"description"`
	Tags        string `gorm:"size:256;not null;default:''" json:"tags"`
	Status      int8   `gorm:"not null;default:1" json:"status"`
}

func (Strategy) TableName() string { return "strategies" }

// StrategyIndicator 策略的量化因子规则组（JSONB）。
// Conditions 形如 {logic, rules[], groups[]}，由规则引擎解析执行。
type StrategyIndicator struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	StrategyID uint      `gorm:"index;not null" json:"strategy_id"`
	Conditions JSON      `gorm:"type:jsonb" json:"conditions"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (StrategyIndicator) TableName() string { return "strategy_indicators" }

// StrategySkill 策略 skill.md（含版本，每次保存 version+1）。
type StrategySkill struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	StrategyID uint      `gorm:"index;not null" json:"strategy_id"`
	Content    string    `gorm:"type:text;not null;default:''" json:"content"`
	Version    int       `gorm:"not null;default:1" json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (StrategySkill) TableName() string { return "strategy_skills" }

// StrategyScreenResult 量化粗筛任务/结果（M2，异步状态机）。
type StrategyScreenResult struct {
	ID            uint       `gorm:"primarykey" json:"id"`
	UserID        uint       `gorm:"index;not null" json:"user_id"`
	StrategyID    uint       `gorm:"index;not null" json:"strategy_id"`
	Status        int8       `gorm:"not null;default:0" json:"status"`
	Params        JSON       `gorm:"type:jsonb" json:"params"`
	UniverseCount int        `gorm:"not null;default:0" json:"universe_count"`
	MatchedCount  int        `gorm:"not null;default:0" json:"matched_count"`
	Candidates    JSON       `gorm:"type:jsonb" json:"candidates"`
	TradeDate     *time.Time `gorm:"type:date" json:"trade_date,omitempty"`
	ErrorMsg      string     `gorm:"size:512;not null;default:''" json:"error_msg"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (StrategyScreenResult) TableName() string { return "strategy_screen_results" }
