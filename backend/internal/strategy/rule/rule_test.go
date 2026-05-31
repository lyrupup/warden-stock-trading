package rule_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"warden/internal/model"
	"warden/internal/strategy/factor"
	"warden/internal/strategy/rule"
)

func closes(vals ...float64) factor.Series {
	bars := make([]model.Kline, len(vals))
	for i, v := range vals {
		d := decimal.NewFromFloat(v)
		bars[i] = model.Kline{Open: d, High: d, Low: d, Close: d, Volume: decimal.NewFromInt(1000)}
	}
	return factor.Series{Bars: bars}
}

// 递增 20 日序列：MA5>MA10>MA20 多头排列成立，且 close>MA5。
func bullSeries() factor.Series {
	return closes(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
}

func TestEval_And_AllMatch(t *testing.T) {
	g := rule.RuleGroup{Logic: "and", Rules: []rule.Rule{
		{Left: factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}, Direction: "bull"}, Op: "==", Right: factor.Expr{Type: "const", Value: true}},
		{Left: factor.Expr{Type: "field", Field: "close"}, Op: ">", Right: factor.Expr{Type: "ma", Period: 5}},
	}}
	res, err := rule.Eval(g, bullSeries())
	assert.NoError(t, err)
	assert.True(t, res.Matched)
	assert.Equal(t, 2, res.TotalLeaves)
	assert.Equal(t, 2, res.HitLeaves)
	assert.Equal(t, []int{0, 1}, res.HitRules)
	assert.InDelta(t, 1.0, res.Score(), 1e-9)
	// 快照含 ma5。
	_, ok := res.Snapshot["ma5"]
	assert.True(t, ok)
}

func TestEval_And_OneFail(t *testing.T) {
	g := rule.RuleGroup{Logic: "and", Rules: []rule.Rule{
		{Left: factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}, Direction: "bull"}, Op: "==", Right: factor.Expr{Type: "const", Value: true}},
		{Left: factor.Expr{Type: "bias", Period: 5}, Op: "<=", Right: factor.Expr{Type: "const", Value: float64(1)}}, // 乖离过大不满足
	}}
	res, err := rule.Eval(g, bullSeries())
	assert.NoError(t, err)
	assert.False(t, res.Matched)
	assert.Equal(t, 1, res.HitLeaves)
	assert.InDelta(t, 0.5, res.Score(), 1e-9)
}

func TestEval_Or(t *testing.T) {
	g := rule.RuleGroup{Logic: "or", Rules: []rule.Rule{
		{Left: factor.Expr{Type: "bias", Period: 5}, Op: "<=", Right: factor.Expr{Type: "const", Value: float64(1)}},                                  // 假
		{Left: factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}, Direction: "bull"}, Op: "==", Right: factor.Expr{Type: "const", Value: true}}, // 真
	}}
	res, err := rule.Eval(g, bullSeries())
	assert.NoError(t, err)
	assert.True(t, res.Matched)
}

func TestEval_NestedGroups(t *testing.T) {
	// (close>MA5) AND ( bias5<=1 OR ma_align bull )
	g := rule.RuleGroup{Logic: "and",
		Rules: []rule.Rule{
			{Left: factor.Expr{Type: "field", Field: "close"}, Op: ">", Right: factor.Expr{Type: "ma", Period: 5}},
		},
		Groups: []rule.RuleGroup{
			{Logic: "or", Rules: []rule.Rule{
				{Left: factor.Expr{Type: "bias", Period: 5}, Op: "<=", Right: factor.Expr{Type: "const", Value: float64(1)}},
				{Left: factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}, Direction: "bull"}, Op: "==", Right: factor.Expr{Type: "const", Value: true}},
			}},
		},
	}
	res, err := rule.Eval(g, bullSeries())
	assert.NoError(t, err)
	assert.True(t, res.Matched)
	assert.Equal(t, 3, res.TotalLeaves)
}

func TestEval_InsufficientDataNotMatched(t *testing.T) {
	// MA60 无法计算 → 该规则不命中，但不报错。
	g := rule.RuleGroup{Logic: "and", Rules: []rule.Rule{
		{Left: factor.Expr{Type: "ma_align", Periods: []int{20, 30, 60}, Direction: "bull"}, Op: "==", Right: factor.Expr{Type: "const", Value: true}},
	}}
	res, err := rule.Eval(g, bullSeries()) // 仅 20 根
	assert.NoError(t, err)
	assert.False(t, res.Matched)
}

func TestEval_InvalidFactorErrors(t *testing.T) {
	g := rule.RuleGroup{Logic: "and", Rules: []rule.Rule{
		{Left: factor.Expr{Type: "rsi"}, Op: ">", Right: factor.Expr{Type: "const", Value: float64(1)}},
	}}
	_, err := rule.Eval(g, bullSeries())
	assert.ErrorIs(t, err, factor.ErrUnknownFactor)
}

func TestEval_BoolVsNumberInvalid(t *testing.T) {
	g := rule.RuleGroup{Logic: "and", Rules: []rule.Rule{
		{Left: factor.Expr{Type: "ma_align", Periods: []int{5, 10}, Direction: "bull"}, Op: ">", Right: factor.Expr{Type: "const", Value: float64(1)}},
	}}
	_, err := rule.Eval(g, bullSeries())
	assert.ErrorIs(t, err, factor.ErrInvalidParam)
}
