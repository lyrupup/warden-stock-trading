// Package catalog 提供 M2 量化因子目录与内置策略模板（只读元数据）。
//
// 因子目录供前端指标构造器动态渲染可选因子与参数；策略模板供「一键创建」。
// 二者是前端构造器与后端引擎口径一致的单一事实源（见 BACKEND.md §5 M2）。
package catalog

import (
	"warden/internal/strategy/factor"
	"warden/internal/strategy/rule"
)

// ParamSpec 因子参数定义。
type ParamSpec struct {
	Key      string `json:"key"`
	Type     string `json:"type"` // int / int[] / float / enum / bool
	Required bool   `json:"required"`
	Default  any    `json:"default,omitempty"`
	Enum     []any  `json:"enum,omitempty"`
	Desc     string `json:"desc"`
}

// Item 因子目录项。
type Item struct {
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	ValueType   string      `json:"value_type"` // number / bool
	Unit        string      `json:"unit,omitempty"`
	Description string      `json:"description"`
	Params      []ParamSpec `json:"params"`
}

// Template 内置策略模板。
type Template struct {
	Key         string         `json:"key"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Tags        string         `json:"tags"`
	Scenario    string         `json:"scenario"`
	Indicators  rule.RuleGroup `json:"indicators"`
}

// Items 返回内置因子目录。
func Items() []Item {
	return []Item{
		{
			Type: "ma", Name: "移动均线 MA", ValueType: "number", Description: "N 日简单移动均线",
			Params: []ParamSpec{{Key: "period", Type: "int", Required: true, Default: 5, Desc: "均线周期"}},
		},
		{
			Type: "ma_align", Name: "均线多头/空头排列", ValueType: "bool",
			Description: "按有序周期校验严格递减(多头)/递增(空头)，如 MA5>MA10>MA20",
			Params: []ParamSpec{
				{Key: "periods", Type: "int[]", Required: true, Default: []int{5, 10, 20}, Desc: "有序均线周期组"},
				{Key: "direction", Type: "enum", Required: true, Default: "bull", Enum: []any{"bull", "bear"}, Desc: "bull 多头 / bear 空头"},
				{Key: "eps", Type: "float", Required: false, Default: 0, Desc: "比较容差，默认 0"},
			},
		},
		{
			Type: "bias", Name: "乖离率 BIAS", ValueType: "number", Unit: "%",
			Description: "(收盘价 - MA(N)) / MA(N) × 100%，衡量偏离均线幅度",
			Params:      []ParamSpec{{Key: "period", Type: "int", Required: true, Default: 5, Desc: "均线周期"}},
		},
		{
			Type: "amplitude", Name: "当日振幅", ValueType: "number", Unit: "%",
			Description: "(最高 - 最低) / 前收盘 × 100%，表征当日多空分歧剧烈程度",
			Params:      []ParamSpec{},
		},
		{
			Type: "amplitude_streak", Name: "连续振幅", ValueType: "bool",
			Description: "连续满足 振幅≥阈值 的天数 ≥ days，捕捉持续剧烈震荡",
			Params: []ParamSpec{
				{Key: "threshold", Type: "float", Required: true, Default: 5, Desc: "振幅阈值(%)"},
				{Key: "days", Type: "int", Required: true, Default: 3, Desc: "连续天数"},
			},
		},
		{
			Type: "pct_change", Name: "N 日涨跌幅", ValueType: "number", Unit: "%",
			Description: "(close_t - close_{t-N}) / close_{t-N} × 100%",
			Params:      []ParamSpec{{Key: "period", Type: "int", Required: true, Default: 5, Desc: "回看天数"}},
		},
		{
			Type: "vol_ratio", Name: "量比", ValueType: "number",
			Description: "当日成交量 / 最近 N 日均量",
			Params:      []ParamSpec{{Key: "period", Type: "int", Required: true, Default: 5, Desc: "均量天数"}},
		},
		{
			Type: "field", Name: "原始字段", ValueType: "number",
			Description: "直接取当前 K 线字段",
			Params: []ParamSpec{{Key: "field", Type: "enum", Required: true, Default: "close",
				Enum: []any{"close", "open", "high", "low", "volume", "amount", "prev_close", "change_percent"}, Desc: "K 线字段"}},
		},
	}
}

// Templates 返回内置策略模板（对应 PRD M2 两个预置策略）。
func Templates() []Template {
	return []Template{
		{
			Key:         "short_ma_bull",
			Name:        "短线均线多头排列",
			Description: "MA5>MA10>MA20 多头排列、价站上 MA5、MA5 乖离≤8%，主升浪追高票。",
			Tags:        "短线,趋势,主升浪",
			Scenario:    "短期趋势强势主升浪，适合追高入场、破短期趋势止损离场。",
			Indicators: rule.RuleGroup{Logic: "and", Rules: []rule.Rule{
				{Left: factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}, Direction: "bull"}, Op: rule.OpEQ, Right: factor.Expr{Type: "const", Value: true}},
				{Left: factor.Expr{Type: "field", Field: "close"}, Op: rule.OpGT, Right: factor.Expr{Type: "ma", Period: 5}},
				{Left: factor.Expr{Type: "bias", Period: 5}, Op: rule.OpLE, Right: factor.Expr{Type: "const", Value: float64(8)}},
			}},
		},
		{
			Key:         "midlong_swing",
			Name:        "中长线多头+短期剧烈震荡",
			Description: "MA20>MA30>MA60 多头排列，且振幅≥5%连续≥3日，中期震荡波段/做 T 票。",
			Tags:        "中长线,震荡,波段",
			Scenario:    "中期趋势向上但短期分歧剧烈，适合波段与日内 T。",
			Indicators: rule.RuleGroup{Logic: "and", Rules: []rule.Rule{
				{Left: factor.Expr{Type: "ma_align", Periods: []int{20, 30, 60}, Direction: "bull"}, Op: rule.OpEQ, Right: factor.Expr{Type: "const", Value: true}},
				{Left: factor.Expr{Type: "amplitude_streak", Threshold: 5, Days: 3}, Op: rule.OpEQ, Right: factor.Expr{Type: "const", Value: true}},
			}},
		},
	}
}
