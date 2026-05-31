// Package market 是 M1 行情外部集成层。
//
// 通过 IMarketProvider 接口屏蔽底层数据源差异（见 BACKEND.md §5 M1）。
// 主力源为 bensema/gotdx（直连通达信，连接池借还）；为保证在无 gotdx 依赖、
// 无外部网络时也能编译与测试，默认提供 stubProvider 返回示例数据。
//
// 注入真实 gotdx：见 gotdx_provider.go（构建标签 `gotdx`），并将 module 的
// Go 版本提升至 gotdx 要求的版本后 `go build -tags gotdx ./...`。
package market

import (
	"context"

	"warden/internal/model"
)

// IMarketProvider 行情数据源抽象。Service 仅依赖此接口，便于切换数据源与单测 mock。
type IMarketProvider interface {
	// Indices 返回当日大盘指数列表。
	Indices(ctx context.Context) ([]model.IndexQuote, error)
	// Quotes 返回给定股票代码的实时/准实时快照行情。
	Quotes(ctx context.Context, codes []string) ([]model.StockQuote, error)
	// Kline 返回 K 线（period: day/week/month/分钟级；adjust: ""/qfq/hfq）。
	Kline(ctx context.Context, code, period, adjust string) ([]model.Kline, error)
	// Search 按代码/名称关键字搜索股票。
	Search(ctx context.Context, kw string) ([]model.StockBrief, error)
}

// newGotdxProvider 由 gotdx_provider.go（构建标签 gotdx）在 init 中注入。
// 默认构建下为 nil，工厂会回退到 stub 实现。
var newGotdxProvider func(maxConn int) IMarketProvider

// NewProvider 根据配置选择行情源实现。
//
//	name="gotdx" 且已用 -tags gotdx 构建时使用真实通达信源；
//	否则回退 stub（返回示例数据，保证可编译可测试）。
func NewProvider(name string, maxConn int) IMarketProvider {
	switch name {
	case "gotdx":
		if newGotdxProvider != nil {
			return newGotdxProvider(maxConn)
		}
		// gotdx 未编译进来，降级 stub。
		return NewStubProvider()
	default:
		return NewStubProvider()
	}
}
