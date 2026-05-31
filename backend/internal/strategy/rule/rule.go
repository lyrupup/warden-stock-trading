// Package rule 是 M2 量化粗筛的规则引擎（见 BACKEND.md §5 M2）。
//
// 规则统一为「left(因子) op right(因子或常量)」，可在 and/or 组内嵌套，
// 表达任意复合条件。规则引擎与策略指标定义、前端构造器共用同一套 schema。
package rule

import (
	"errors"
	"fmt"

	"warden/internal/strategy/factor"
)

// 支持的比较操作符。
const (
	OpGT = ">"
	OpGE = ">="
	OpLT = "<"
	OpLE = "<="
	OpEQ = "=="
	OpNE = "!="
)

// Rule 单条规则：left op right。布尔因子默认与 {const:true} 比较。
type Rule struct {
	Left  factor.Expr `json:"left"`
	Op    string      `json:"op"`
	Right factor.Expr `json:"right"`
}

// RuleGroup 规则组，支持 and/or 与嵌套子组。
type RuleGroup struct {
	Logic  string      `json:"logic"`
	Rules  []Rule      `json:"rules"`
	Groups []RuleGroup `json:"groups,omitempty"`
}

// Result 规则组求值结果。
type Result struct {
	Matched     bool                    // 整组是否命中
	HitRules    []int                   // 顶层 Rules 中命中的规则下标（供前端高亮）
	TotalLeaves int                     // 递归叶子规则总数（评分分母）
	HitLeaves   int                     // 递归命中叶子数（评分分子）
	Snapshot    map[string]factor.Value // 本次用到的因子快照值
}

// Score 命中规则占比（默认评分口径，见 BACKEND.md §5 M2）。无叶子时为 0。
func (r Result) Score() float64 {
	if r.TotalLeaves == 0 {
		return 0
	}
	return float64(r.HitLeaves) / float64(r.TotalLeaves)
}

// Eval 对一只股票（K 线序列）执行规则组求值。
// 因子数据不足（ErrInsufficientData）视为该规则不命中，不中断整轮；
// 未知因子 / 参数非法等配置错误向上暴露。
func Eval(g RuleGroup, s factor.Series) (Result, error) {
	snap := make(map[string]factor.Value)
	res := Result{Snapshot: snap}
	matched, err := evalGroup(g, s, &res, true)
	if err != nil {
		return Result{}, err
	}
	res.Matched = matched
	return res, nil
}

func evalGroup(g RuleGroup, s factor.Series, res *Result, top bool) (bool, error) {
	logic := g.Logic
	if logic == "" {
		logic = "and"
	}
	if logic != "and" && logic != "or" {
		return false, fmt.Errorf("%w: 未知逻辑 %q", factor.ErrInvalidParam, logic)
	}

	matched := logic == "and" // and 初始真、or 初始假
	hasChild := false

	for i, r := range g.Rules {
		hasChild = true
		res.TotalLeaves++
		ok, err := evalRule(r, s, res.Snapshot)
		if err != nil {
			return false, err
		}
		if ok {
			res.HitLeaves++
			if top {
				res.HitRules = append(res.HitRules, i)
			}
		}
		if logic == "and" {
			matched = matched && ok
		} else {
			matched = matched || ok
		}
	}

	for _, sub := range g.Groups {
		hasChild = true
		ok, err := evalGroup(sub, s, res, false)
		if err != nil {
			return false, err
		}
		if logic == "and" {
			matched = matched && ok
		} else {
			matched = matched || ok
		}
	}

	if !hasChild {
		return false, nil // 空组视为不命中
	}
	return matched, nil
}

// evalRule 计算单条规则。因子数据不足时按「不命中」处理（返回 false, nil）。
func evalRule(r Rule, s factor.Series, snap map[string]factor.Value) (bool, error) {
	lv, err := factor.Eval(r.Left, s)
	if err != nil {
		if errors.Is(err, factor.ErrInsufficientData) {
			return false, nil
		}
		return false, err
	}
	if k := factor.Key(r.Left); k != "" {
		snap[k] = lv
	}

	rv, err := factor.Eval(r.Right, s)
	if err != nil {
		if errors.Is(err, factor.ErrInsufficientData) {
			return false, nil
		}
		return false, err
	}
	if k := factor.Key(r.Right); k != "" {
		snap[k] = rv
	}

	return compare(lv, r.Op, rv)
}

func compare(l factor.Value, op string, r factor.Value) (bool, error) {
	// 布尔比较：任一为布尔则两者都按布尔处理，仅支持 == / !=。
	if l.IsBool || r.IsBool {
		if l.IsBool != r.IsBool {
			return false, fmt.Errorf("%w: 布尔因子与数值因子不可比较", factor.ErrInvalidParam)
		}
		switch op {
		case OpEQ:
			return l.Bool == r.Bool, nil
		case OpNE:
			return l.Bool != r.Bool, nil
		default:
			return false, fmt.Errorf("%w: 布尔因子不支持操作符 %q", factor.ErrInvalidParam, op)
		}
	}

	cmp := l.Num.Cmp(r.Num) // -1 / 0 / 1
	switch op {
	case OpGT:
		return cmp > 0, nil
	case OpGE:
		return cmp >= 0, nil
	case OpLT:
		return cmp < 0, nil
	case OpLE:
		return cmp <= 0, nil
	case OpEQ:
		return cmp == 0, nil
	case OpNE:
		return cmp != 0, nil
	default:
		return false, fmt.Errorf("%w: 未知操作符 %q", factor.ErrInvalidParam, op)
	}
}
