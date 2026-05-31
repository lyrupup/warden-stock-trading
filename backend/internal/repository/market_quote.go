package repository

import (
	"context"

	"gorm.io/gorm"

	"warden/internal/model"
)

// MarketQuoteRepository 行情快照仓储。快照表为公共数据（不归属单一用户），
// 用作外部源失败时的降级兜底（见 BACKEND.md §5 M1 降级）。
type MarketQuoteRepository interface {
	SaveStockQuotes(ctx context.Context, quotes []model.StockQuote) error
	// LatestStockQuotes 返回给定代码各自最新一条快照。
	LatestStockQuotes(ctx context.Context, codes []string) ([]model.StockQuote, error)
	SaveIndexQuotes(ctx context.Context, quotes []model.IndexQuote) error
	LatestIndexQuotes(ctx context.Context) ([]model.IndexQuote, error)
}

type marketQuoteRepo struct {
	db *gorm.DB
}

// NewMarketQuoteRepository 构造基于 GORM 的行情快照仓储。
func NewMarketQuoteRepository(db *gorm.DB) MarketQuoteRepository {
	return &marketQuoteRepo{db: db}
}

func (r *marketQuoteRepo) SaveStockQuotes(ctx context.Context, quotes []model.StockQuote) error {
	if r.db == nil || len(quotes) == 0 {
		return nil // 无 DB 时快照兜底为 best-effort，静默跳过
	}
	return r.db.WithContext(ctx).Create(&quotes).Error
}

func (r *marketQuoteRepo) LatestStockQuotes(ctx context.Context, codes []string) ([]model.StockQuote, error) {
	if r.db == nil {
		return nil, errNoDB
	}
	if len(codes) == 0 {
		return nil, nil
	}
	// 取每个 stock_code 的最新一条快照（按 snapshot_at 倒序去重）。
	var quotes []model.StockQuote
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (stock_code) *
			FROM stock_quotes
			WHERE stock_code IN ?
			ORDER BY stock_code, snapshot_at DESC
		`, codes).
		Scan(&quotes).Error
	return quotes, err
}

func (r *marketQuoteRepo) SaveIndexQuotes(ctx context.Context, quotes []model.IndexQuote) error {
	if r.db == nil || len(quotes) == 0 {
		return nil // 无 DB 时快照兜底为 best-effort，静默跳过
	}
	return r.db.WithContext(ctx).Create(&quotes).Error
}

func (r *marketQuoteRepo) LatestIndexQuotes(ctx context.Context) ([]model.IndexQuote, error) {
	if r.db == nil {
		return nil, errNoDB
	}
	var quotes []model.IndexQuote
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (index_code) *
			FROM index_quotes
			ORDER BY index_code, snapshot_at DESC
		`).
		Scan(&quotes).Error
	return quotes, err
}
