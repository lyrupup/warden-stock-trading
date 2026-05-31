package request

import "warden/internal/strategy/rule"

// CreateStrategyReq 新建策略请求。
type CreateStrategyReq struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Tags        string          `json:"tags"`
	Indicators  *rule.RuleGroup `json:"indicators"`
}

// UpdateStrategyReq 更新策略请求。
type UpdateStrategyReq struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Tags        string          `json:"tags"`
	Indicators  *rule.RuleGroup `json:"indicators"`
}

// SaveSkillReq 保存 skill.md 请求（保存即生成新版本）。
type SaveSkillReq struct {
	Content string `json:"content" binding:"required"`
}

// Universe 粗筛股票池范围。
type Universe struct {
	Type  string   `json:"type"`  // all / board / watchlist / codes
	Board string   `json:"board"` // type=board 时给板块代码
	Codes []string `json:"codes"` // type=codes 时给代码列表
}

// ScreenReq 量化粗筛请求。
type ScreenReq struct {
	Universe  Universe `json:"universe"`
	TradeDate string   `json:"trade_date"` // 留空=最新交易日
	Limit     int      `json:"limit"`      // 候选上限（按评分截断）
}

// PreviewScreenReq 同步快速粗筛预览请求（携带临时指标定义，不落任务表）。
type PreviewScreenReq struct {
	ScreenReq
	Indicators *rule.RuleGroup `json:"indicators" binding:"required"`
}
