package market

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"warden/internal/model"
)

// stubProvider 返回确定性的示例行情数据，用于本地开发与单元测试。
//
// TODO(gotdx): 接入真实 gotdx provider（见 gotdx_provider.go，构建标签 gotdx）。
// 真实数据源就绪后，将 config 的 market.provider 设为 gotdx 并以 -tags gotdx 构建即可，
// Service/Handler 无需改动（依赖 IMarketProvider 接口）。
type stubProvider struct{}

// NewStubProvider 返回示例数据 provider（占位实现）。
func NewStubProvider() IMarketProvider { return &stubProvider{} }

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func (p *stubProvider) Indices(ctx context.Context) ([]model.IndexQuote, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	today := time.Now()
	return []model.IndexQuote{
		{IndexCode: "000001", IndexName: "上证指数", Price: dec("3120.50"), ChangeAmount: dec("12.30"), ChangePercent: dec("0.40"), Volume: dec("280000000"), Amount: dec("356000000000"), TradeDate: today, SnapshotAt: today},
		{IndexCode: "399001", IndexName: "深证成指", Price: dec("9850.20"), ChangeAmount: dec("-23.10"), ChangePercent: dec("-0.23"), Volume: dec("310000000"), Amount: dec("420000000000"), TradeDate: today, SnapshotAt: today},
		{IndexCode: "399006", IndexName: "创业板指", Price: dec("1920.80"), ChangeAmount: dec("5.60"), ChangePercent: dec("0.29"), Volume: dec("120000000"), Amount: dec("180000000000"), TradeDate: today, SnapshotAt: today},
		{IndexCode: "000688", IndexName: "科创50", Price: dec("880.10"), ChangeAmount: dec("3.20"), ChangePercent: dec("0.37"), Volume: dec("60000000"), Amount: dec("90000000000"), TradeDate: today, SnapshotAt: today},
		{IndexCode: "000300", IndexName: "沪深300", Price: dec("3650.40"), ChangeAmount: dec("8.90"), ChangePercent: dec("0.24"), Volume: dec("150000000"), Amount: dec("260000000000"), TradeDate: today, SnapshotAt: today},
	}, nil
}

func (p *stubProvider) Quotes(ctx context.Context, codes []string) ([]model.StockQuote, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	today := time.Now()
	out := make([]model.StockQuote, 0, len(codes))
	for _, code := range codes {
		out = append(out, model.StockQuote{
			StockCode:     code,
			StockName:     "示例股票" + code,
			Price:         dec("10.50"),
			Open:          dec("10.20"),
			High:          dec("10.80"),
			Low:           dec("10.10"),
			PrevClose:     dec("10.30"),
			ChangePercent: dec("1.94"),
			Volume:        dec("12000000"),
			Amount:        dec("126000000"),
			TurnoverRate:  dec("2.35"),
			TradeDate:     today,
			SnapshotAt:    today,
		})
	}
	return out, nil
}

func (p *stubProvider) Kline(ctx context.Context, code, period, adjust string) ([]model.Kline, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	base := time.Now().AddDate(0, 0, -4)
	out := make([]model.Kline, 0, 5)
	for i := 0; i < 5; i++ {
		out = append(out, model.Kline{
			Date:   base.AddDate(0, 0, i).Format("2006-01-02"),
			Open:   dec("10.00"),
			High:   dec("10.50"),
			Low:    dec("9.90"),
			Close:  dec("10.30"),
			Volume: dec("11000000"),
			Amount: dec("113000000"),
		})
	}
	return out, nil
}

func (p *stubProvider) Search(ctx context.Context, kw string) ([]model.StockBrief, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []model.StockBrief{
		{StockCode: "600519", StockName: "贵州茅台", Market: "SH"},
		{StockCode: "000858", StockName: "五粮液", Market: "SZ"},
	}, nil
}
