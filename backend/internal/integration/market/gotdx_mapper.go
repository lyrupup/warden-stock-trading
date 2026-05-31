//go:build gotdx

// 本文件仅在 `-tags gotdx` 构建时编译。
//
// 集中放置 gotdx 原始返回（proto.*）-> warden model 的字段映射与编码转换。
// 通达信协议价格已在 gotdx 内部按 /100（行情）或 /1000（K 线）还原成「元」，
// 成交量单位为「手」，成交额单位为「元」。
package market

import (
	"strings"
	"time"

	"github.com/bensema/gotdx/proto"
	"github.com/bensema/gotdx/types"
	"github.com/shopspring/decimal"

	"warden/internal/model"
)

// indexRef 描述一个需要展示的大盘指数及其在通达信中的市场/代码。
type indexRef struct {
	market uint8
	code   string
	name   string
}

// trackedIndices 与 stub_provider 保持一致的指数集合：上证/深成/创业板/科创50/沪深300。
var trackedIndices = []indexRef{
	{uint8(types.MarketSH), "000001", "上证指数"},
	{uint8(types.MarketSZ), "399001", "深证成指"},
	{uint8(types.MarketSZ), "399006", "创业板指"},
	{uint8(types.MarketSH), "000688", "科创50"},
	{uint8(types.MarketSH), "000300", "沪深300"},
}

// marketOf 按代码前缀推断通达信市场代码（沪 1 / 深 0 / 北 2）。
func marketOf(code string) uint8 {
	if code == "" {
		return uint8(types.MarketSZ)
	}
	switch code[0] {
	case '6', '5', '9': // 沪市A股/沪市ETF(5)/沪市B股(9)
		return uint8(types.MarketSH)
	case '4', '8': // 北交所
		return uint8(types.MarketBJ)
	default: // 0/3 深市A股/创业板，1 深市ETF/可转债
		return uint8(types.MarketSZ)
	}
}

func marketsOf(codes []string) []uint8 {
	out := make([]uint8, len(codes))
	for i, c := range codes {
		out[i] = marketOf(c)
	}
	return out
}

func marketName(m uint8) string { return types.Market(m).String() }

// periodOf 把 API 的 period 字符串映射为 gotdx 的 K 线类别枚举。
func periodOf(period string) uint16 {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "week", "w", "1w":
		return proto.KLINE_TYPE_WEEKLY
	case "month", "1mon", "mon":
		return proto.KLINE_TYPE_MONTHLY
	case "1m", "1min":
		return proto.KLINE_TYPE_1MIN
	case "5m", "5min":
		return proto.KLINE_TYPE_5MIN
	case "15m", "15min":
		return proto.KLINE_TYPE_15MIN
	case "30m", "30min":
		return proto.KLINE_TYPE_30MIN
	case "60m", "1h", "hour":
		return proto.KLINE_TYPE_1HOUR
	default: // "", day, d, 1d
		return proto.KLINE_TYPE_DAILY
	}
}

// adjustOf 把 API 的复权口径映射为 gotdx 复权枚举（服务端复权）。
func adjustOf(adjust string) uint16 {
	switch strings.ToLower(strings.TrimSpace(adjust)) {
	case "qfq":
		return types.AdjustQFQ
	case "hfq":
		return types.AdjustHFQ
	default:
		return types.AdjustNone
	}
}

func fromFloat(v float64) decimal.Decimal { return decimal.NewFromFloat(v) }

// mapIndex 把指数概况映射为 model.IndexQuote。reply.Diff 即「当前-昨收」涨跌值。
func mapIndex(ref indexRef, r *proto.GetIndexInfoReply, now time.Time) model.IndexQuote {
	changePct := 0.0
	if r.PreClose != 0 {
		changePct = r.Diff / r.PreClose * 100
	}
	return model.IndexQuote{
		IndexCode:     ref.code,
		IndexName:     ref.name,
		Price:         fromFloat(r.Close),
		ChangeAmount:  fromFloat(r.Diff),
		ChangePercent: fromFloat(changePct).Round(4),
		Volume:        decimal.NewFromInt(int64(r.Vol)),
		Amount:        fromFloat(r.Amount),
		TradeDate:     now,
		SnapshotAt:    now,
	}
}

// mapQuote 把五档行情明细映射为 model.StockQuote。name 由调用方按证券列表补齐（可为空）。
func mapQuote(q proto.SecurityQuote, name string, now time.Time) model.StockQuote {
	changePct := 0.0
	if q.PreClose != 0 {
		changePct = (q.Price - q.PreClose) / q.PreClose * 100
	}
	return model.StockQuote{
		StockCode:     q.Code,
		StockName:     name,
		Price:         fromFloat(q.Price),
		Open:          fromFloat(q.Open),
		High:          fromFloat(q.High),
		Low:           fromFloat(q.Low),
		PrevClose:     fromFloat(q.PreClose),
		ChangePercent: fromFloat(changePct).Round(4),
		Volume:        decimal.NewFromInt(int64(q.Vol)),
		Amount:        fromFloat(q.Amount),
		// gotdx 的 Turnover = Vol(手) × 10000 / 流通股(万股)，量纲上恰为「真实换手率% × 10000」。
		// 故除以 10000 还原为百分比；流通股拉取失败时 gotdx 置 0，除后仍为 0。
		TurnoverRate: fromFloat(q.Turnover / 10000).Round(4),
		TradeDate:    now,
		SnapshotAt:   now,
	}
}

// mapKline 把单根 K 线映射为 model.Kline。日/周/月线只取日期，分钟级附带时间。
func mapKline(b proto.SecurityBar, intraday bool) model.Kline {
	layout := "2006-01-02"
	if intraday {
		layout = "2006-01-02 15:04"
	}
	return model.Kline{
		Date:   b.DateTime.Format(layout),
		Open:   fromFloat(b.Open),
		High:   fromFloat(b.High),
		Low:    fromFloat(b.Low),
		Close:  fromFloat(b.Close),
		Volume: fromFloat(b.Vol),
		Amount: fromFloat(b.Amount),
	}
}

func isIntraday(period uint16) bool {
	switch period {
	case proto.KLINE_TYPE_DAILY, proto.KLINE_TYPE_WEEKLY, proto.KLINE_TYPE_MONTHLY:
		return false
	default:
		return true
	}
}
