package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"warden/internal/dto/request"
	"warden/internal/mock"
	"warden/internal/model"
	"warden/internal/service"
	"warden/pkg/errcode"
)

const testUserID uint = 1

func quote(code string) model.StockQuote {
	return model.StockQuote{StockCode: code, StockName: "n" + code, Price: decimal.RequireFromString("10.50")}
}

// TestMarketService_WatchlistQuotes 覆盖 M1 行情竖切核心范式：
// 缓存命中（不调外部）、provider 调用（回写缓存+快照）、provider 失败降级 stale。
func TestMarketService_WatchlistQuotes(t *testing.T) {
	codes := []string{"600519", "000858"}
	items := []model.WatchlistItem{
		{UserID: testUserID, StockCode: "600519"},
		{UserID: testUserID, StockCode: "000858"},
	}

	tests := []struct {
		name      string
		setup     func(p *mock.MockIMarketProvider, wr *mock.MockWatchlistRepository, qr *mock.MockMarketQuoteRepository, ca *mock.MockQuoteCache)
		wantLen   int
		wantStale bool
		wantErr   bool
	}{
		{
			name: "缓存命中-不调用外部源",
			setup: func(p *mock.MockIMarketProvider, wr *mock.MockWatchlistRepository, qr *mock.MockMarketQuoteRepository, ca *mock.MockQuoteCache) {
				wr.EXPECT().List(gomock.Any(), testUserID).Return(items, nil)
				ca.EXPECT().GetQuotes(gomock.Any(), codes).Return([]model.StockQuote{quote("600519"), quote("000858")}, []string{}, nil)
				// provider.Quotes 不应被调用（缓存全命中）。
			},
			wantLen:   2,
			wantStale: false,
		},
		{
			name: "缓存未命中-调用provider并回写",
			setup: func(p *mock.MockIMarketProvider, wr *mock.MockWatchlistRepository, qr *mock.MockMarketQuoteRepository, ca *mock.MockQuoteCache) {
				wr.EXPECT().List(gomock.Any(), testUserID).Return(items, nil)
				ca.EXPECT().GetQuotes(gomock.Any(), codes).Return(nil, codes, nil)
				fresh := []model.StockQuote{quote("600519"), quote("000858")}
				p.EXPECT().Quotes(gomock.Any(), codes).Return(fresh, nil)
				ca.EXPECT().SetQuotes(gomock.Any(), fresh).Return(nil)
				qr.EXPECT().SaveStockQuotes(gomock.Any(), fresh).Return(nil)
			},
			wantLen:   2,
			wantStale: false,
		},
		{
			name: "provider失败-降级返回快照并标记stale",
			setup: func(p *mock.MockIMarketProvider, wr *mock.MockWatchlistRepository, qr *mock.MockMarketQuoteRepository, ca *mock.MockQuoteCache) {
				wr.EXPECT().List(gomock.Any(), testUserID).Return(items, nil)
				ca.EXPECT().GetQuotes(gomock.Any(), codes).Return(nil, codes, nil)
				p.EXPECT().Quotes(gomock.Any(), codes).Return(nil, errors.New("provider down"))
				qr.EXPECT().LatestStockQuotes(gomock.Any(), codes).
					Return([]model.StockQuote{quote("600519"), quote("000858")}, nil)
				// 降级路径不回写缓存。
			},
			wantLen:   2,
			wantStale: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			provider := mock.NewMockIMarketProvider(ctrl)
			watchRepo := mock.NewMockWatchlistRepository(ctrl)
			quoteRepo := mock.NewMockMarketQuoteRepository(ctrl)
			cache := mock.NewMockQuoteCache(ctrl)
			tt.setup(provider, watchRepo, quoteRepo, cache)

			svc := service.NewMarketService(provider, watchRepo, quoteRepo, cache)
			got, err := svc.WatchlistQuotes(context.Background(), testUserID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
			for _, q := range got {
				assert.Equal(t, tt.wantStale, q.Stale)
			}
		})
	}
}

// TestMarketService_WatchlistQuotes_EmptyWatchlist 空自选不应调用外部源。
func TestMarketService_WatchlistQuotes_EmptyWatchlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := mock.NewMockIMarketProvider(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	quoteRepo := mock.NewMockMarketQuoteRepository(ctrl)
	cache := mock.NewMockQuoteCache(ctrl)

	watchRepo.EXPECT().List(gomock.Any(), testUserID).Return([]model.WatchlistItem{}, nil)

	svc := service.NewMarketService(provider, watchRepo, quoteRepo, cache)
	got, err := svc.WatchlistQuotes(context.Background(), testUserID)
	assert.NoError(t, err)
	assert.Empty(t, got)
}

// TestMarketService_WatchlistQuotes_NoCache cache 为 nil 时直连 provider。
func TestMarketService_WatchlistQuotes_NoCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	codes := []string{"600519"}
	provider := mock.NewMockIMarketProvider(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	quoteRepo := mock.NewMockMarketQuoteRepository(ctrl)

	watchRepo.EXPECT().List(gomock.Any(), testUserID).
		Return([]model.WatchlistItem{{StockCode: "600519"}}, nil)
	provider.EXPECT().Quotes(gomock.Any(), codes).Return([]model.StockQuote{quote("600519")}, nil)
	quoteRepo.EXPECT().SaveStockQuotes(gomock.Any(), gomock.Any()).Return(nil)

	svc := service.NewMarketService(provider, watchRepo, quoteRepo, nil)
	got, err := svc.WatchlistQuotes(context.Background(), testUserID)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.False(t, got[0].Stale)
}

// TestMarketService_AddWatchlist_Duplicate 已存在时返回 ErrWatchlistExists，不创建。
func TestMarketService_AddWatchlist_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := mock.NewMockIMarketProvider(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	quoteRepo := mock.NewMockMarketQuoteRepository(ctrl)

	watchRepo.EXPECT().ExistsByCode(gomock.Any(), testUserID, "600519").Return(true, nil)

	svc := service.NewMarketService(provider, watchRepo, quoteRepo, nil)
	_, err := svc.AddWatchlist(context.Background(), testUserID, &request.AddWatchlistReq{StockCode: "600519"})
	assert.ErrorIs(t, err, errcode.ErrWatchlistExists)
}

// TestMarketService_Indices_Degrade provider 失败时降级到最近指数快照。
func TestMarketService_Indices_Degrade(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := mock.NewMockIMarketProvider(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	quoteRepo := mock.NewMockMarketQuoteRepository(ctrl)

	provider.EXPECT().Indices(gomock.Any()).Return(nil, errors.New("down"))
	quoteRepo.EXPECT().LatestIndexQuotes(gomock.Any()).
		Return([]model.IndexQuote{{IndexCode: "000001", IndexName: "上证指数"}}, nil)

	svc := service.NewMarketService(provider, watchRepo, quoteRepo, nil)
	got, err := svc.Indices(context.Background())
	assert.NoError(t, err)
	assert.Len(t, got, 1)
}
