//go:build gotdx

// 本文件仅在 `-tags gotdx` 构建时编译，提供真实通达信行情源实现。
//
// 默认构建（无 gotdx 标签）不包含本文件，NewProvider 回退到 stubProvider。
// 接入：将 module Go 版本提升至 gotdx 要求（≥1.26，go.mod 已锚定），
// 然后 `go build -tags gotdx ./...`。Service/Handler 仅依赖 IMarketProvider，无需改动。
package market

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bensema/gotdx/types"

	"warden/internal/model"
)

// klineMaxCount 单次 K 线请求条数上限。
// 通达信服务端单包上限不足 800（gotdx 内部会额外 +1 取昨收，>800 时返回空包并在
// 解析阶段 panic），实测 480 稳定可用，对日/周/月线展示足够。
const klineMaxCount uint16 = 480

func init() {
	// 注册真实实现，供 provider.go 的 NewProvider("gotdx", ...) 使用。
	newGotdxProvider = func(maxConn int) IMarketProvider {
		p := &gotdxProvider{pool: newGotdxPool(maxConn)}
		// 后台预热全市场证券名索引：个股名补齐与关键字搜索复用，失败不影响其他接口。
		go func() { _ = p.names.ensure(p) }()
		return p
	}
}

type gotdxProvider struct {
	pool  *gotdxPool
	names nameIndex
}

// recoverAs 把 gotdx 解析层可能出现的 panic（畸形/空响应）转成 error，避免拖垮进程。
func recoverAs(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("gotdx panic: %v", r)
	}
}

func (p *gotdxProvider) Indices(ctx context.Context) (out []model.IndexQuote, err error) {
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	cli, err := p.pool.Get()
	if err != nil {
		return nil, err
	}
	broken := true
	defer func() { recoverAs(&err); p.pool.Put(cli, broken) }()

	now := time.Now()
	out = make([]model.IndexQuote, 0, len(trackedIndices))
	for _, ref := range trackedIndices {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		reply, e := cli.StockIndexInfo(ref.market, ref.code)
		if e != nil {
			return nil, e
		}
		out = append(out, mapIndex(ref, reply, now))
	}
	broken = false
	return out, nil
}

func (p *gotdxProvider) Quotes(ctx context.Context, codes []string) (out []model.StockQuote, err error) {
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		return nil, nil
	}
	cli, err := p.pool.Get()
	if err != nil {
		return nil, err
	}
	broken := true
	defer func() { recoverAs(&err); p.pool.Put(cli, broken) }()

	rows, e := cli.StockQuotesDetail(marketsOf(codes), codes)
	if e != nil {
		return nil, e
	}
	broken = false

	now := time.Now()
	out = make([]model.StockQuote, 0, len(rows))
	for _, q := range rows {
		out = append(out, mapQuote(q, p.names.lookup(q.Code), now))
	}
	return out, nil
}

func (p *gotdxProvider) Kline(ctx context.Context, code, period, adjust string) (out []model.Kline, err error) {
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	cli, err := p.pool.Get()
	if err != nil {
		return nil, err
	}
	broken := true
	defer func() { recoverAs(&err); p.pool.Put(cli, broken) }()

	cat := periodOf(period)
	// start=0 取最新；times=1 不合并周期；adjust 由服务端复权（qfq/hfq/none）。
	bars, e := cli.StockKLine(cat, marketOf(code), code, 0, klineMaxCount, 1, adjustOf(adjust))
	if e != nil {
		return nil, e
	}
	broken = false

	intraday := isIntraday(cat)
	out = make([]model.Kline, 0, len(bars))
	for _, b := range bars {
		out = append(out, mapKline(b, intraday))
	}
	return out, nil
}

func (p *gotdxProvider) Search(ctx context.Context, kw string) ([]model.StockBrief, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	kw = strings.TrimSpace(kw)
	if kw == "" {
		return nil, nil
	}
	if err := p.names.ensure(p); err != nil {
		return nil, err
	}
	return p.names.search(kw, 30), nil
}

// nameIndex 懒加载全市场证券「代码 -> 名称/市场」索引。
// 通达信无服务端关键字搜索，需本地拉全量列表后过滤；加载一次后进程内缓存。
type nameIndex struct {
	mu       sync.RWMutex
	loaded   bool
	byCode   map[string]string // code -> name
	marketOf map[string]uint8  // code -> market
	order    []string          // 稳定排序用的代码列表
}

func (ni *nameIndex) lookup(code string) string {
	ni.mu.RLock()
	defer ni.mu.RUnlock()
	return ni.byCode[code]
}

// ensure 确保全量列表已加载；失败不缓存，允许下次重试。
func (ni *nameIndex) ensure(p *gotdxProvider) (err error) {
	ni.mu.RLock()
	done := ni.loaded
	ni.mu.RUnlock()
	if done {
		return nil
	}

	cli, err := p.pool.Get()
	if err != nil {
		return err
	}
	broken := true
	defer func() { recoverAs(&err); p.pool.Put(cli, broken) }()

	byCode := make(map[string]string, 12000)
	mkt := make(map[string]uint8, 12000)
	order := make([]string, 0, 12000)
	for _, m := range []uint8{uint8(types.MarketSH), uint8(types.MarketSZ), uint8(types.MarketBJ)} {
		list, e := cli.StockAll(m)
		if e != nil {
			return e
		}
		for _, s := range list {
			code := strings.TrimSpace(s.Code)
			if code == "" {
				continue
			}
			if _, ok := byCode[code]; !ok {
				order = append(order, code)
			}
			byCode[code] = s.Name
			mkt[code] = m
		}
	}
	broken = false

	ni.mu.Lock()
	ni.byCode, ni.marketOf, ni.order, ni.loaded = byCode, mkt, order, true
	ni.mu.Unlock()
	return nil
}

// search 按代码前缀或名称子串过滤，返回至多 limit 条（代码升序）。
func (ni *nameIndex) search(kw string, limit int) []model.StockBrief {
	ni.mu.RLock()
	defer ni.mu.RUnlock()

	kwLower := strings.ToLower(kw)
	hits := make([]model.StockBrief, 0, limit)
	for _, code := range ni.order {
		name := ni.byCode[code]
		if strings.HasPrefix(code, kw) || strings.Contains(strings.ToLower(name), kwLower) {
			hits = append(hits, model.StockBrief{
				StockCode: code,
				StockName: name,
				Market:    marketName(ni.marketOf[code]),
			})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].StockCode < hits[j].StockCode })
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits
}
