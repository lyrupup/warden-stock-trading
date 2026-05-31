package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"warden/internal/dto/request"
	"warden/internal/mock"
	"warden/internal/model"
	"warden/internal/service"
	"warden/internal/strategy/factor"
	"warden/internal/strategy/rule"
	"warden/pkg/errcode"
)

// bullKlines 生成一段递增 K 线，满足 MA5>MA10>MA20 多头排列且 close>MA5。
func bullKlines(n int) []model.Kline {
	bars := make([]model.Kline, n)
	for i := 0; i < n; i++ {
		v := decimal.NewFromInt(int64(i + 1))
		bars[i] = model.Kline{Open: v, High: v, Low: v, Close: v, Volume: decimal.NewFromInt(1000)}
	}
	return bars
}

// flatKlines 生成一段横盘 K 线（均线相等，不构成多头排列）。
func flatKlines(n int) []model.Kline {
	bars := make([]model.Kline, n)
	for i := 0; i < n; i++ {
		v := decimal.NewFromInt(10)
		bars[i] = model.Kline{Open: v, High: v, Low: v, Close: v, Volume: decimal.NewFromInt(1000)}
	}
	return bars
}

func bullStrategyGroup() *rule.RuleGroup {
	return &rule.RuleGroup{Logic: "and", Rules: []rule.Rule{
		{Left: factor.Expr{Type: "ma_align", Periods: []int{5, 10, 20}, Direction: "bull"}, Op: "==", Right: factor.Expr{Type: "const", Value: true}},
		{Left: factor.Expr{Type: "field", Field: "close"}, Op: ">", Right: factor.Expr{Type: "ma", Period: 5}},
	}}
}

func indicatorModel(g *rule.RuleGroup) *model.StrategyIndicator {
	b, _ := json.Marshal(g)
	return &model.StrategyIndicator{StrategyID: 7, Conditions: model.JSON(b)}
}

// TestStrategyService_RunScreen 覆盖粗筛核心：命中的票进入候选，不命中的被过滤。
func TestStrategyService_RunScreen(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockStrategyRepository(ctrl)
	screenRepo := mock.NewMockScreenResultRepository(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	provider := mock.NewMockIMarketProvider(ctrl)

	const uid, sid = uint(1), uint(7)
	repo.EXPECT().GetByID(gomock.Any(), uid, sid).Return(&model.Strategy{BaseModel: model.BaseModel{ID: sid}, UserID: uid}, nil)
	repo.EXPECT().GetIndicator(gomock.Any(), sid).Return(indicatorModel(bullStrategyGroup()), nil)

	// 创建任务记录（回填 ID）。
	screenRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, r *model.StrategyScreenResult) error {
		r.ID = 99
		return nil
	})

	// 自定义代码池：600519 多头命中，000001 横盘不命中。
	provider.EXPECT().Kline(gomock.Any(), "600519", "day", "qfq").Return(bullKlines(30), nil)
	provider.EXPECT().Kline(gomock.Any(), "000001", "day", "qfq").Return(flatKlines(30), nil)
	// 候选补名（600519）。
	provider.EXPECT().Quotes(gomock.Any(), gomock.Any()).Return([]model.StockQuote{{StockCode: "600519", StockName: "贵州茅台"}}, nil)

	var saved *model.StrategyScreenResult
	screenRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, r *model.StrategyScreenResult) error {
		saved = r
		return nil
	})

	svc := service.NewStrategyService(repo, screenRepo, watchRepo, provider)
	taskID, err := svc.RunScreen(context.Background(), uid, sid, &request.ScreenReq{
		Universe: request.Universe{Type: "codes", Codes: []string{"600519", "000001"}},
	})
	assert.NoError(t, err)
	assert.Equal(t, uint(99), taskID)
	assert.NotNil(t, saved)
	assert.Equal(t, model.ScreenStatusDone, saved.Status)
	assert.Equal(t, 2, saved.UniverseCount)
	assert.Equal(t, 1, saved.MatchedCount)

	var cands []map[string]any
	assert.NoError(t, json.Unmarshal(saved.Candidates, &cands))
	assert.Len(t, cands, 1)
	assert.Equal(t, "600519", cands[0]["stock_code"])
	assert.Equal(t, "贵州茅台", cands[0]["stock_name"])
}

// TestStrategyService_RunScreen_NoIndicator 未定义指标时报错。
func TestStrategyService_RunScreen_NoIndicator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockStrategyRepository(ctrl)
	screenRepo := mock.NewMockScreenResultRepository(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	provider := mock.NewMockIMarketProvider(ctrl)

	const uid, sid = uint(1), uint(7)
	repo.EXPECT().GetByID(gomock.Any(), uid, sid).Return(&model.Strategy{BaseModel: model.BaseModel{ID: sid}}, nil)
	repo.EXPECT().GetIndicator(gomock.Any(), sid).Return(nil, nil)

	svc := service.NewStrategyService(repo, screenRepo, watchRepo, provider)
	_, err := svc.RunScreen(context.Background(), uid, sid, &request.ScreenReq{Universe: request.Universe{Type: "codes", Codes: []string{"600519"}}})
	assert.ErrorIs(t, err, errcode.ErrIndicatorInvalid)
}

// TestStrategyService_PreviewScreen 预览使用临时指标，watchlist 池。
func TestStrategyService_PreviewScreen(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockStrategyRepository(ctrl)
	screenRepo := mock.NewMockScreenResultRepository(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	provider := mock.NewMockIMarketProvider(ctrl)

	watchRepo.EXPECT().List(gomock.Any(), uint(1)).Return([]model.WatchlistItem{{StockCode: "600519", StockName: "贵州茅台"}}, nil)
	provider.EXPECT().Kline(gomock.Any(), "600519", "day", "qfq").Return(bullKlines(30), nil)

	svc := service.NewStrategyService(repo, screenRepo, watchRepo, provider)
	res, err := svc.PreviewScreen(context.Background(), 1, &request.PreviewScreenReq{
		ScreenReq:  request.ScreenReq{Universe: request.Universe{Type: "watchlist"}},
		Indicators: bullStrategyGroup(),
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, res.MatchedCount)
	assert.Len(t, res.Candidates, 1)
	assert.Equal(t, "600519", res.Candidates[0].StockCode)
	assert.Equal(t, "贵州茅台", res.Candidates[0].StockName)
}

// TestStrategyService_PreviewScreen_InvalidIndicator 未知因子报指标非法。
func TestStrategyService_PreviewScreen_InvalidIndicator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockStrategyRepository(ctrl)
	screenRepo := mock.NewMockScreenResultRepository(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	provider := mock.NewMockIMarketProvider(ctrl)

	svc := service.NewStrategyService(repo, screenRepo, watchRepo, provider)
	_, err := svc.PreviewScreen(context.Background(), 1, &request.PreviewScreenReq{
		ScreenReq: request.ScreenReq{Universe: request.Universe{Type: "codes", Codes: []string{"600519"}}},
		Indicators: &rule.RuleGroup{Logic: "and", Rules: []rule.Rule{
			{Left: factor.Expr{Type: "rsi"}, Op: ">", Right: factor.Expr{Type: "const", Value: float64(1)}},
		}},
	})
	assert.ErrorIs(t, err, errcode.ErrIndicatorInvalid)
}

// TestStrategyService_PreviewScreen_Concurrent 并发拉 K 线时，候选聚合不丢、不串台、不死锁。
// 池中 5 命中 + 3 不命中，限制并发度 = 2，结果应为 5 个候选，且评分按 desc 排序。
func TestStrategyService_PreviewScreen_Concurrent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockStrategyRepository(ctrl)
	screenRepo := mock.NewMockScreenResultRepository(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	provider := mock.NewMockIMarketProvider(ctrl)

	hit := []string{"600519", "601318", "000858", "002594", "300750"}
	miss := []string{"000001", "000002", "000063"}
	for _, c := range hit {
		provider.EXPECT().Kline(gomock.Any(), c, "day", "qfq").Return(bullKlines(30), nil)
	}
	for _, c := range miss {
		provider.EXPECT().Kline(gomock.Any(), c, "day", "qfq").Return(flatKlines(30), nil)
	}
	provider.EXPECT().Quotes(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	codes := append(append([]string{}, hit...), miss...)
	svc := service.NewStrategyService(repo, screenRepo, watchRepo, provider, service.WithScreenConcurrency(2))
	res, err := svc.PreviewScreen(context.Background(), 1, &request.PreviewScreenReq{
		ScreenReq:  request.ScreenReq{Universe: request.Universe{Type: "codes", Codes: codes}},
		Indicators: bullStrategyGroup(),
	})
	assert.NoError(t, err)
	assert.Equal(t, len(codes), res.UniverseCount)
	assert.Equal(t, 5, res.MatchedCount)
	assert.Len(t, res.Candidates, 5)
	for i := 1; i < len(res.Candidates); i++ {
		assert.True(t, res.Candidates[i-1].Score.GreaterThanOrEqual(res.Candidates[i].Score), "候选应按 score 降序")
	}
}

// TestStrategyService_Screen_UnsupportedUniverse 全市场池暂不支持。
func TestStrategyService_Screen_UnsupportedUniverse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock.NewMockStrategyRepository(ctrl)
	screenRepo := mock.NewMockScreenResultRepository(ctrl)
	watchRepo := mock.NewMockWatchlistRepository(ctrl)
	provider := mock.NewMockIMarketProvider(ctrl)

	svc := service.NewStrategyService(repo, screenRepo, watchRepo, provider)
	_, err := svc.PreviewScreen(context.Background(), 1, &request.PreviewScreenReq{
		ScreenReq:  request.ScreenReq{Universe: request.Universe{Type: "all"}},
		Indicators: bullStrategyGroup(),
	})
	assert.ErrorIs(t, err, errcode.ErrUniverseInvalid)
}
