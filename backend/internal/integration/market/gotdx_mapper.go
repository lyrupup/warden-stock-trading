//go:build gotdx

// 本文件仅在 `-tags gotdx` 构建时编译。
//
// 这里集中放置 gotdx 原始返回 -> warden model 的字段映射与复权计算。
// 由于真实 gotdx 的结构体/方法签名以其版本为准，接入时需按实际类型调整下列适配函数，
// 但对外仍只暴露 IMarketProvider 接口。applyAdjust 为纯函数，须有单测覆盖
// （前复权/后复权/不复权口径，见 BACKEND.md §6.2）。
package market

import (
	"github.com/bensema/gotdx"

	"warden/internal/model"
)

// marketOf 推断单只股票所属市场代码（如沪 1 / 深 0），需按 gotdx 约定实现。
func marketOf(code string) int {
	// TODO(gotdx): 按代码前缀映射通达信 market id。
	if len(code) > 0 && (code[0] == '6') {
		return 1 // 上交所
	}
	return 0 // 深交所
}

func marketsOf(codes []string) []int {
	out := make([]int, len(codes))
	for i, c := range codes {
		out[i] = marketOf(c)
	}
	return out
}

func symbolsOf(codes []string) []string { return codes }

// periodOf 把 API 的 period 字符串转换为 gotdx 的 K 线类型枚举。
func periodOf(period string) int {
	// TODO(gotdx): 映射 day/week/month/1m... 到 gotdx KType。
	return 0
}

func mapIndices(rows interface{}) []model.IndexQuote {
	// TODO(gotdx): 按真实返回结构映射。
	_ = rows
	return nil
}

func mapQuotes(rows interface{}) []model.StockQuote {
	// TODO(gotdx): 按真实返回结构映射。
	_ = rows
	return nil
}

func mapKlines(bars interface{}) []model.Kline {
	// TODO(gotdx): 按真实返回结构映射。
	_ = bars
	return nil
}

func mapBriefs(rows interface{}) []model.StockBrief {
	// TODO(gotdx): 按真实返回结构映射。
	_ = rows
	return nil
}

// applyAdjust 按除权除息事件逐根 K 线计算前/后复权（adjust: ""/qfq/hfq）。
func applyAdjust(klines []model.Kline, xdxr []gotdx.XDXRInfo, adjust string) []model.Kline {
	// TODO(gotdx): 按复权因子换算（见 BACKEND.md §5 复权计算说明）。
	_ = xdxr
	_ = adjust
	return klines
}
