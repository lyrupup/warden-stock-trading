// Package response 定义出参 DTO。
package response

import "warden/internal/model"

// WatchItem 自选股列表项（对齐 openapi WatchItem）。
type WatchItem struct {
	ID        uint   `json:"id"`
	StockCode string `json:"stock_code"`
	StockName string `json:"stock_name"`
	GroupName string `json:"group_name"`
	Remark    string `json:"remark"`
}

// FromWatchlistItem 将 model 转换为列表项 DTO。
func FromWatchlistItem(m model.WatchlistItem) WatchItem {
	return WatchItem{
		ID:        m.ID,
		StockCode: m.StockCode,
		StockName: m.StockName,
		GroupName: m.GroupName,
		Remark:    m.Remark,
	}
}

// FromWatchlistItems 批量转换。
func FromWatchlistItems(items []model.WatchlistItem) []WatchItem {
	out := make([]WatchItem, 0, len(items))
	for _, it := range items {
		out = append(out, FromWatchlistItem(it))
	}
	return out
}
