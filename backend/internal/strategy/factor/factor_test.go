package factor_test

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"warden/internal/model"
	"warden/internal/strategy/factor"
)

func bar(close, high, low, vol float64) model.Kline {
	return model.Kline{
		Open:   decimal.NewFromFloat(close),
		High:   decimal.NewFromFloat(high),
		Low:    decimal.NewFromFloat(low),
		Close:  decimal.NewFromFloat(close),
		Volume: decimal.NewFromFloat(vol),
		Amount: decimal.NewFromFloat(close * vol),
	}
}

// closes 构造一段只关心收盘价的 K 线序列（high/low=close，便于均线/乖离测试）。
func closes(vals ...float64) factor.Series {
	bars := make([]model.Kline, len(vals))
	for i, v := range vals {
		bars[i] = bar(v, v, v, 1000)
	}
	return factor.Series{Bars: bars}
}

func TestMA(t *testing.T) {
	s := closes(10, 11, 12, 13, 14) // MA5 = 12
	ma, err := factor.MA(s, 5)
	assert.NoError(t, err)
	assert.Equal(t, "12", ma.String())

	ma3, err := factor.MA(s, 3) // (12+13+14)/3 = 13
	assert.NoError(t, err)
	assert.Equal(t, "13", ma3.String())

	_, err = factor.MA(s, 6) // 数据不足
	assert.ErrorIs(t, err, factor.ErrInsufficientData)
}

func TestEval_MaAlign(t *testing.T) {
	// 递增序列 → 短均线高于长均线 → 多头排列成立。
	bull := closes(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	v, err := factor.Eval(factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}, Direction: "bull"}, bull)
	assert.NoError(t, err)
	assert.True(t, v.IsBool)
	assert.True(t, v.Bool)

	// 同一序列判空头排列应为假。
	v, err = factor.Eval(factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}, Direction: "bear"}, bull)
	assert.NoError(t, err)
	assert.False(t, v.Bool)

	// 递减序列 → 空头排列成立。
	bear := closes(20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1)
	v, err = factor.Eval(factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}, Direction: "bear"}, bear)
	assert.NoError(t, err)
	assert.True(t, v.Bool)

	// 周期不足报数据不足。
	_, err = factor.Eval(factor.Expr{Type: "ma_align", Periods: []int{20, 30, 60}, Direction: "bull"}, bull)
	assert.ErrorIs(t, err, factor.ErrInsufficientData)

	// 周期数不足 2 个为参数非法。
	_, err = factor.Eval(factor.Expr{Type: "ma_align", Periods: []int{5}}, bull)
	assert.ErrorIs(t, err, factor.ErrInvalidParam)
}

func TestEval_Bias(t *testing.T) {
	// MA5=12, close=14 → bias = (14-12)/12*100 ≈ 16.6667%
	s := closes(10, 11, 12, 13, 14)
	v, err := factor.Eval(factor.Expr{Type: "bias", Period: 5}, s)
	assert.NoError(t, err)
	assert.Equal(t, "16.6667", v.Num.Round(4).String())
}

func TestEval_Amplitude(t *testing.T) {
	// 前收=10，当日 high=12 low=9 → 振幅=(12-9)/10*100=30%
	s := factor.Series{Bars: []model.Kline{bar(10, 10, 10, 1000), bar(11, 12, 9, 1000)}}
	v, err := factor.Eval(factor.Expr{Type: "amplitude"}, s)
	assert.NoError(t, err)
	assert.Equal(t, "30", v.Num.String())

	// 只有一根 bar 无前收 → 数据不足。
	_, err = factor.Eval(factor.Expr{Type: "amplitude"}, factor.Series{Bars: []model.Kline{bar(10, 12, 9, 1)}})
	assert.ErrorIs(t, err, factor.ErrInsufficientData)
}

func TestEval_AmplitudeStreak(t *testing.T) {
	// 构造最近 3 日振幅均 >= 5%，更早一日振幅小。
	bars := []model.Kline{
		bar(10, 10, 10, 1000), // 基准前收
		bar(10, 10.2, 9.9, 1), // 振幅 (0.3)/10=3% < 5
		bar(10, 11, 9, 1),     // (2)/10=20% >=5
		bar(10, 10.6, 9.8, 1), // (0.8)/10=8% >=5
		bar(10, 11, 9.4, 1),   // (1.6)/10=16% >=5
	}
	s := factor.Series{Bars: bars}
	v, err := factor.Eval(factor.Expr{Type: "amplitude_streak", Threshold: 5, Days: 3}, s)
	assert.NoError(t, err)
	assert.True(t, v.Bool)

	// 要求连续 4 日则不满足（第 4 日往前是 3% ）。
	v, err = factor.Eval(factor.Expr{Type: "amplitude_streak", Threshold: 5, Days: 4}, s)
	assert.NoError(t, err)
	assert.False(t, v.Bool)
}

func TestEval_PctChange(t *testing.T) {
	// close[t]=12, close[t-3]=10 → (12-10)/10*100=20%
	s := closes(9, 10, 10.5, 11, 12)
	v, err := factor.Eval(factor.Expr{Type: "pct_change", Period: 3}, s)
	assert.NoError(t, err)
	assert.Equal(t, "20", v.Num.Round(4).String())
}

func TestEval_Field(t *testing.T) {
	s := factor.Series{Bars: []model.Kline{bar(10, 10, 10, 500), bar(12, 13, 9, 800)}}
	v, _ := factor.Eval(factor.Expr{Type: "field", Field: "close"}, s)
	assert.Equal(t, "12", v.Num.String())
	v, _ = factor.Eval(factor.Expr{Type: "field", Field: "prev_close"}, s)
	assert.Equal(t, "10", v.Num.String())
	v, _ = factor.Eval(factor.Expr{Type: "field", Field: "volume"}, s)
	assert.Equal(t, "800", v.Num.String())
	_, err := factor.Eval(factor.Expr{Type: "field", Field: "unknown"}, s)
	assert.ErrorIs(t, err, factor.ErrInvalidParam)
}

func TestEval_Const(t *testing.T) {
	v, _ := factor.Eval(factor.Expr{Type: "const", Value: true}, factor.Series{})
	assert.True(t, v.IsBool)
	assert.True(t, v.Bool)
	v, _ = factor.Eval(factor.Expr{Type: "const", Value: float64(8)}, factor.Series{})
	assert.False(t, v.IsBool)
	assert.Equal(t, "8", v.Num.String())
}

func TestEval_UnknownFactor(t *testing.T) {
	_, err := factor.Eval(factor.Expr{Type: "rsi"}, closes(1, 2, 3))
	assert.ErrorIs(t, err, factor.ErrUnknownFactor)
}

func TestKey(t *testing.T) {
	assert.Equal(t, "ma5", factor.Key(factor.Expr{Type: "ma", Period: 5}))
	assert.Equal(t, "bias10", factor.Key(factor.Expr{Type: "bias", Period: 10}))
	assert.Equal(t, "amplitude", factor.Key(factor.Expr{Type: "amplitude"}))
	assert.Equal(t, "ma_align_5_10_20", factor.Key(factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}}))
	assert.Equal(t, "close", factor.Key(factor.Expr{Type: "field", Field: "close"}))
	assert.Equal(t, "", factor.Key(factor.Expr{Type: "const", Value: 1}))
}

func TestValue_Decimal(t *testing.T) {
	assert.Equal(t, "1", factor.Value{Bool: true, IsBool: true}.Decimal().String())
	assert.Equal(t, "0", factor.Value{Bool: false, IsBool: true}.Decimal().String())
	assert.True(t, errors.Is(factor.ErrInsufficientData, factor.ErrInsufficientData))
}
