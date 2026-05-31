// Package factor 是 M2 量化指标的因子引擎（见 BACKEND.md §5 M2）。
//
// 把盘面经验量化为可计算因子：均线(ma)、均线多头/空头排列(ma_align)、
// 乖离率(bias)、当日振幅(amplitude)、连续振幅(amplitude_streak)、
// N 日涨跌幅(pct_change)、原始字段(field)、量比(vol_ratio)、常量(const)。
//
// 所有计算均为纯函数，基于前复权日 K 线序列（升序，最后一根为当前 bar），
// 便于 TDD 覆盖。因子可作为规则的左/右操作数，与常量或另一个因子比较。
package factor

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	"warden/internal/model"
)

// 因子计算可能返回的错误。ErrInsufficientData 由规则引擎吞掉视为「不命中」，
// 其余（未知因子/参数非法）应作为配置错误向上暴露。
var (
	ErrInsufficientData = errors.New("factor: K线数据不足以计算该因子")
	ErrUnknownFactor    = errors.New("factor: 未知因子类型")
	ErrInvalidParam     = errors.New("factor: 因子参数非法")
)

// Value 因子值：数值或布尔（布尔因子如 ma_align / amplitude_streak）。
type Value struct {
	Num    decimal.Decimal
	Bool   bool
	IsBool bool
}

// Decimal 把因子值统一表示为 decimal（布尔 → 1/0），用于候选快照展示。
func (v Value) Decimal() decimal.Decimal {
	if v.IsBool {
		if v.Bool {
			return decimal.NewFromInt(1)
		}
		return decimal.Zero
	}
	return v.Num
}

// Series 升序日 K 线序列，Bars[len-1] 为当前 bar。
type Series struct {
	Bars []model.Kline
}

// Expr 因子表达式，与 JSONB 中规则的 left/right 对齐。
type Expr struct {
	Type      string  `json:"type"`
	Period    int     `json:"period,omitempty"`
	Periods   []int   `json:"periods,omitempty"`
	Direction string  `json:"direction,omitempty"`
	Field     string  `json:"field,omitempty"`
	Threshold float64 `json:"threshold,omitempty"`
	Days      int     `json:"days,omitempty"`
	Eps       float64 `json:"eps,omitempty"`
	Value     any     `json:"value,omitempty"`
}

const hundred = 100

// Eval 按 type 分发计算因子值。
func Eval(e Expr, s Series) (Value, error) {
	switch e.Type {
	case "ma":
		ma, err := MA(s, e.Period)
		if err != nil {
			return Value{}, err
		}
		return Value{Num: ma}, nil
	case "ma_align":
		return evalMaAlign(e, s)
	case "bias":
		return evalBias(e, s)
	case "amplitude":
		return evalAmplitude(s)
	case "amplitude_streak":
		return evalAmplitudeStreak(e, s)
	case "pct_change":
		return evalPctChange(e, s)
	case "field":
		return evalField(e, s)
	case "vol_ratio":
		return evalVolRatio(e, s)
	case "const":
		return evalConst(e)
	default:
		return Value{}, fmt.Errorf("%w: %q", ErrUnknownFactor, e.Type)
	}
}

// MA 计算最近 period 根收盘价的简单移动均线。
func MA(s Series, period int) (decimal.Decimal, error) {
	if period <= 0 {
		return decimal.Zero, fmt.Errorf("%w: ma period=%d", ErrInvalidParam, period)
	}
	n := len(s.Bars)
	if n < period {
		return decimal.Zero, ErrInsufficientData
	}
	sum := decimal.Zero
	for i := n - period; i < n; i++ {
		sum = sum.Add(s.Bars[i].Close)
	}
	return sum.Div(decimal.NewFromInt(int64(period))), nil
}

func evalMaAlign(e Expr, s Series) (Value, error) {
	if len(e.Periods) < 2 {
		return Value{}, fmt.Errorf("%w: ma_align 需至少 2 个周期", ErrInvalidParam)
	}
	dir := e.Direction
	if dir == "" {
		dir = "bull"
	}
	if dir != "bull" && dir != "bear" {
		return Value{}, fmt.Errorf("%w: ma_align direction=%q", ErrInvalidParam, dir)
	}
	mas := make([]decimal.Decimal, len(e.Periods))
	for i, p := range e.Periods {
		ma, err := MA(s, p)
		if err != nil {
			return Value{}, err
		}
		mas[i] = ma
	}
	eps := decimal.NewFromFloat(e.Eps)
	ok := true
	for i := 0; i+1 < len(mas); i++ {
		if dir == "bull" {
			// 多头：MA(p_i) > MA(p_{i+1}) + eps，严格递减。
			if !mas[i].GreaterThan(mas[i+1].Add(eps)) {
				ok = false
				break
			}
		} else {
			// 空头：MA(p_i) < MA(p_{i+1}) - eps，严格递增。
			if !mas[i].LessThan(mas[i+1].Sub(eps)) {
				ok = false
				break
			}
		}
	}
	return Value{Bool: ok, IsBool: true}, nil
}

func evalBias(e Expr, s Series) (Value, error) {
	ma, err := MA(s, e.Period)
	if err != nil {
		return Value{}, err
	}
	if ma.IsZero() {
		return Value{}, fmt.Errorf("%w: bias MA=0", ErrInvalidParam)
	}
	close := s.Bars[len(s.Bars)-1].Close
	bias := close.Sub(ma).Div(ma).Mul(decimal.NewFromInt(hundred))
	return Value{Num: bias}, nil
}

// amplitudeAt 计算第 i 根 bar 的振幅% = (high-low)/prev_close*100，需 i>=1。
func amplitudeAt(s Series, i int) (decimal.Decimal, error) {
	if i < 1 || i >= len(s.Bars) {
		return decimal.Zero, ErrInsufficientData
	}
	prevClose := s.Bars[i-1].Close
	if prevClose.IsZero() {
		return decimal.Zero, fmt.Errorf("%w: amplitude prev_close=0", ErrInvalidParam)
	}
	bar := s.Bars[i]
	return bar.High.Sub(bar.Low).Div(prevClose).Mul(decimal.NewFromInt(hundred)), nil
}

func evalAmplitude(s Series) (Value, error) {
	amp, err := amplitudeAt(s, len(s.Bars)-1)
	if err != nil {
		return Value{}, err
	}
	return Value{Num: amp}, nil
}

func evalAmplitudeStreak(e Expr, s Series) (Value, error) {
	if e.Days <= 0 {
		return Value{}, fmt.Errorf("%w: amplitude_streak days=%d", ErrInvalidParam, e.Days)
	}
	threshold := decimal.NewFromFloat(e.Threshold)
	streak := 0
	for i := len(s.Bars) - 1; i >= 1; i-- {
		amp, err := amplitudeAt(s, i)
		if err != nil {
			break
		}
		if amp.GreaterThanOrEqual(threshold) {
			streak++
		} else {
			break
		}
	}
	return Value{Bool: streak >= e.Days, IsBool: true}, nil
}

func evalPctChange(e Expr, s Series) (Value, error) {
	if e.Period <= 0 {
		return Value{}, fmt.Errorf("%w: pct_change period=%d", ErrInvalidParam, e.Period)
	}
	n := len(s.Bars)
	if n < e.Period+1 {
		return Value{}, ErrInsufficientData
	}
	cur := s.Bars[n-1].Close
	base := s.Bars[n-1-e.Period].Close
	if base.IsZero() {
		return Value{}, fmt.Errorf("%w: pct_change base=0", ErrInvalidParam)
	}
	pct := cur.Sub(base).Div(base).Mul(decimal.NewFromInt(hundred))
	return Value{Num: pct}, nil
}

func evalField(e Expr, s Series) (Value, error) {
	n := len(s.Bars)
	if n == 0 {
		return Value{}, ErrInsufficientData
	}
	bar := s.Bars[n-1]
	switch e.Field {
	case "close":
		return Value{Num: bar.Close}, nil
	case "open":
		return Value{Num: bar.Open}, nil
	case "high":
		return Value{Num: bar.High}, nil
	case "low":
		return Value{Num: bar.Low}, nil
	case "volume":
		return Value{Num: bar.Volume}, nil
	case "amount":
		return Value{Num: bar.Amount}, nil
	case "prev_close":
		if n < 2 {
			return Value{}, ErrInsufficientData
		}
		return Value{Num: s.Bars[n-2].Close}, nil
	case "change_percent":
		if n < 2 {
			return Value{}, ErrInsufficientData
		}
		prev := s.Bars[n-2].Close
		if prev.IsZero() {
			return Value{}, fmt.Errorf("%w: change_percent prev=0", ErrInvalidParam)
		}
		pct := bar.Close.Sub(prev).Div(prev).Mul(decimal.NewFromInt(hundred))
		return Value{Num: pct}, nil
	default:
		return Value{}, fmt.Errorf("%w: field=%q", ErrInvalidParam, e.Field)
	}
}

func evalVolRatio(e Expr, s Series) (Value, error) {
	if e.Period <= 0 {
		return Value{}, fmt.Errorf("%w: vol_ratio period=%d", ErrInvalidParam, e.Period)
	}
	n := len(s.Bars)
	if n < e.Period+1 {
		return Value{}, ErrInsufficientData
	}
	sum := decimal.Zero
	for i := n - 1 - e.Period; i < n-1; i++ {
		sum = sum.Add(s.Bars[i].Volume)
	}
	avg := sum.Div(decimal.NewFromInt(int64(e.Period)))
	if avg.IsZero() {
		return Value{}, fmt.Errorf("%w: vol_ratio avg=0", ErrInvalidParam)
	}
	return Value{Num: s.Bars[n-1].Volume.Div(avg)}, nil
}

func evalConst(e Expr) (Value, error) {
	switch v := e.Value.(type) {
	case bool:
		return Value{Bool: v, IsBool: true}, nil
	case float64:
		return Value{Num: decimal.NewFromFloat(v)}, nil
	case int:
		return Value{Num: decimal.NewFromInt(int64(v))}, nil
	case int64:
		return Value{Num: decimal.NewFromInt(v)}, nil
	case string:
		d, err := decimal.NewFromString(v)
		if err != nil {
			return Value{}, fmt.Errorf("%w: const value=%q", ErrInvalidParam, v)
		}
		return Value{Num: d}, nil
	case nil:
		return Value{}, fmt.Errorf("%w: const value 缺失", ErrInvalidParam)
	default:
		return Value{}, fmt.Errorf("%w: const value 类型不支持", ErrInvalidParam)
	}
}

// Key 返回因子在候选快照中的稳定键名（const 因子返回空串，不入快照）。
func Key(e Expr) string {
	switch e.Type {
	case "ma", "bias", "pct_change", "vol_ratio":
		return e.Type + strconv.Itoa(e.Period)
	case "amplitude":
		return "amplitude"
	case "amplitude_streak":
		return "amplitude_streak"
	case "ma_align":
		ps := make([]string, len(e.Periods))
		for i, p := range e.Periods {
			ps[i] = strconv.Itoa(p)
		}
		return "ma_align_" + strings.Join(ps, "_")
	case "field":
		return e.Field
	default:
		return ""
	}
}
