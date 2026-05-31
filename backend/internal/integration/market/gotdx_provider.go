//go:build gotdx

// 本文件仅在 `-tags gotdx` 构建时编译，提供真实通达信行情源实现。
//
// 注意：真实 gotdx API 的方法名/签名可能与下方脚手架不完全一致，接入时以「能编译」
// 为准，用 mapXxx / 适配函数封装差异，但务必保持 IMarketProvider 接口不变
// （见 BACKEND.md §5）。
package market

import (
	"context"

	"github.com/bensema/gotdx"

	"warden/internal/model"
)

func init() {
	// 注册真实实现，供 provider.go 的 NewProvider("gotdx", ...) 使用。
	newGotdxProvider = func(maxConn int) IMarketProvider {
		return &gotdxProvider{pool: newGotdxPool(maxConn)}
	}
}

type gotdxProvider struct{ pool *gotdxPool }

func (p *gotdxProvider) Indices(ctx context.Context) ([]model.IndexQuote, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cli, err := p.pool.Get()
	if err != nil {
		return nil, err
	}
	rows, err := cli.IndexQuotes() // TODO: 以真实 gotdx API 适配
	p.pool.Put(cli, err != nil)
	if err != nil {
		return nil, err
	}
	return mapIndices(rows), nil
}

func (p *gotdxProvider) Quotes(ctx context.Context, codes []string) ([]model.StockQuote, error) {
	if err := ctx.Err(); err != nil {
		return nil, err // 传播超时/取消
	}
	cli, err := p.pool.Get()
	if err != nil {
		return nil, err
	}
	rows, err := cli.StockQuotesDetail(marketsOf(codes), symbolsOf(codes))
	p.pool.Put(cli, err != nil) // 出错视为连接可能损坏，归还时丢弃
	if err != nil {
		return nil, err
	}
	return mapQuotes(rows), nil
}

func (p *gotdxProvider) Kline(ctx context.Context, code, period, adjust string) ([]model.Kline, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cli, err := p.pool.Get()
	if err != nil {
		return nil, err
	}
	bars, err := cli.StockKLine(marketOf(code), code, periodOf(period), 0, 800)
	if err != nil {
		p.pool.Put(cli, true)
		return nil, err
	}
	var xdxr []gotdx.XDXRInfo
	if adjust != "" { // 需要复权时拉除权除息数据
		xdxr, err = cli.GetXDXRInfo(marketOf(code), code)
	}
	p.pool.Put(cli, err != nil)
	if err != nil {
		return nil, err
	}
	return applyAdjust(mapKlines(bars), xdxr, adjust), nil // 自算前/后复权
}

func (p *gotdxProvider) Search(ctx context.Context, kw string) ([]model.StockBrief, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cli, err := p.pool.Get()
	if err != nil {
		return nil, err
	}
	rows, err := cli.SecurityList(kw) // TODO: 以真实 gotdx API 适配
	p.pool.Put(cli, err != nil)
	if err != nil {
		return nil, err
	}
	return mapBriefs(rows), nil
}
