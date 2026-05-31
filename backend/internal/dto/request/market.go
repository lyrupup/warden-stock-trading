// Package request 定义入参 DTO（含 validator 校验标签）。
package request

// AddWatchlistReq 添加自选股请求。
type AddWatchlistReq struct {
	StockCode string `json:"stock_code" binding:"required"`
	GroupName string `json:"group_name"`
	Remark    string `json:"remark"`
}
