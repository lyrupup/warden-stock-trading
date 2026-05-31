// Package service 是业务编排层：调用 repository / 外部集成，处理缓存与降级，
// 不处理 HTTP。所有从 context 取 userID，不信任前端传入（见 BACKEND.md §2）。
package service

import (
	"context"

	"warden/internal/dto/request"
	"warden/internal/integration/market"
	"warden/internal/model"
	"warden/internal/repository"
	"warden/pkg/errcode"
)

// QuoteCache 行情缓存抽象（Redis 实现见 cache 适配）。
// 单测以 mock 替换，禁止真实外呼。
type QuoteCache interface {
	// GetQuotes 返回命中的快照与未命中的代码列表。
	GetQuotes(ctx context.Context, codes []string) (hits []model.StockQuote, missing []string, err error)
	// SetQuotes 回写缓存。
	SetQuotes(ctx context.Context, quotes []model.StockQuote) error
}

// MarketService M1 行情业务接口。
type MarketService interface {
	Indices(ctx context.Context) ([]model.IndexQuote, error)
	ListWatchlist(ctx context.Context, userID uint) ([]model.WatchlistItem, error)
	AddWatchlist(ctx context.Context, userID uint, req *request.AddWatchlistReq) (*model.WatchlistItem, error)
	DeleteWatchlist(ctx context.Context, userID, id uint) error
	WatchlistQuotes(ctx context.Context, userID uint) ([]model.StockQuote, error)
	StockDetail(ctx context.Context, code string) (*model.StockQuote, error)
	Kline(ctx context.Context, code, period, adjust string) ([]model.Kline, error)
	Search(ctx context.Context, kw string) ([]model.StockBrief, error)
}

type marketService struct {
	provider  market.IMarketProvider
	watchRepo repository.WatchlistRepository
	quoteRepo repository.MarketQuoteRepository
	cache     QuoteCache // 可为 nil（无 Redis 时降级为直连 provider）
}

// NewMarketService 注入依赖构造行情业务。cache 允许为 nil。
func NewMarketService(
	provider market.IMarketProvider,
	watchRepo repository.WatchlistRepository,
	quoteRepo repository.MarketQuoteRepository,
	cache QuoteCache,
) MarketService {
	return &marketService{provider: provider, watchRepo: watchRepo, quoteRepo: quoteRepo, cache: cache}
}

// Indices 当日大盘指数。外部源失败时降级为最近快照。
func (s *marketService) Indices(ctx context.Context) ([]model.IndexQuote, error) {
	indices, err := s.provider.Indices(ctx)
	if err != nil {
		if snap, sErr := s.quoteRepo.LatestIndexQuotes(ctx); sErr == nil && len(snap) > 0 {
			return snap, nil
		}
		return nil, errcode.ErrMarketProvider.Wrap(err)
	}
	_ = s.quoteRepo.SaveIndexQuotes(ctx, indices) // 快照兜底，best-effort
	return indices, nil
}

func (s *marketService) ListWatchlist(ctx context.Context, userID uint) ([]model.WatchlistItem, error) {
	return s.watchRepo.List(ctx, userID)
}

func (s *marketService) AddWatchlist(ctx context.Context, userID uint, req *request.AddWatchlistReq) (*model.WatchlistItem, error) {
	exists, err := s.watchRepo.ExistsByCode(ctx, userID, req.StockCode)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errcode.ErrWatchlistExists
	}
	group := req.GroupName
	if group == "" {
		group = "default"
	}
	item := &model.WatchlistItem{
		UserID:    userID,
		StockCode: req.StockCode,
		GroupName: group,
		Remark:    req.Remark,
	}
	// best-effort 补全股票名称（失败不阻塞）。
	if quotes, qErr := s.provider.Quotes(ctx, []string{req.StockCode}); qErr == nil && len(quotes) > 0 {
		item.StockName = quotes[0].StockName
	}
	if err := s.watchRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *marketService) DeleteWatchlist(ctx context.Context, userID, id uint) error {
	return s.watchRepo.Delete(ctx, userID, id)
}

// WatchlistQuotes 自选股实时行情：优先缓存 → provider → 快照兜底（降级 stale）。
// 这是 M1 行情竖切的核心范式。
func (s *marketService) WatchlistQuotes(ctx context.Context, userID uint) ([]model.StockQuote, error) {
	items, err := s.watchRepo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []model.StockQuote{}, nil
	}
	codes := make([]string, 0, len(items))
	for _, it := range items {
		codes = append(codes, it.StockCode)
	}
	return s.fetchQuotes(ctx, codes), nil
}

// StockDetail 个股行情详情。
func (s *marketService) StockDetail(ctx context.Context, code string) (*model.StockQuote, error) {
	quotes := s.fetchQuotes(ctx, []string{code})
	if len(quotes) == 0 {
		return nil, errcode.ErrStockNotFound
	}
	return &quotes[0], nil
}

func (s *marketService) Kline(ctx context.Context, code, period, adjust string) ([]model.Kline, error) {
	klines, err := s.provider.Kline(ctx, code, period, adjust)
	if err != nil {
		return nil, errcode.ErrMarketProvider.Wrap(err)
	}
	return klines, nil
}

func (s *marketService) Search(ctx context.Context, kw string) ([]model.StockBrief, error) {
	briefs, err := s.provider.Search(ctx, kw)
	if err != nil {
		return nil, errcode.ErrMarketProvider.Wrap(err)
	}
	return briefs, nil
}

// fetchQuotes 实现「缓存命中 → provider → 快照兜底(stale)」的取数流程，按 codes 顺序返回。
func (s *marketService) fetchQuotes(ctx context.Context, codes []string) []model.StockQuote {
	byCode := make(map[string]model.StockQuote, len(codes))
	missing := codes

	// 1) 缓存命中：命中部分直接采用，未命中部分继续向下取。
	if s.cache != nil {
		hits, miss, cErr := s.cache.GetQuotes(ctx, codes)
		if cErr == nil {
			for _, q := range hits {
				byCode[q.StockCode] = q
			}
			missing = miss
		}
	}

	// 全部命中缓存：不调用外部源。
	if len(missing) > 0 {
		quotes, err := s.provider.Quotes(ctx, missing)
		if err != nil {
			// 2) 降级：外部源失败 → 返回最近快照并标记 stale。
			if snap, sErr := s.quoteRepo.LatestStockQuotes(ctx, missing); sErr == nil {
				for _, q := range snap {
					q.Stale = true
					byCode[q.StockCode] = q
				}
			}
		} else {
			// 3) 命中外部源：回写缓存 + 快照兜底（best-effort）。
			for _, q := range quotes {
				byCode[q.StockCode] = q
			}
			if s.cache != nil {
				_ = s.cache.SetQuotes(ctx, quotes)
			}
			_ = s.quoteRepo.SaveStockQuotes(ctx, quotes)
		}
	}

	// 按入参顺序组装结果（缺失的代码跳过）。
	out := make([]model.StockQuote, 0, len(codes))
	for _, code := range codes {
		if q, ok := byCode[code]; ok {
			out = append(out, q)
		}
	}
	return out
}
