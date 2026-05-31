// Package handler 是 HTTP 适配层：参数绑定/校验、调 Service、组装统一响应。
// 不含业务逻辑、不直连 DB（见 BACKEND.md §1.2）。
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"warden/internal/dto/request"
	"warden/internal/dto/response"
	"warden/internal/middleware"
	"warden/internal/service"
	"warden/pkg/errcode"
	resp "warden/pkg/response"
)

// MarketHandler 处理 M1 行情相关端点（对齐 openapi.yaml /market/*）。
type MarketHandler struct {
	svc service.MarketService
}

// NewMarketHandler 构造行情 Handler。
func NewMarketHandler(svc service.MarketService) *MarketHandler {
	return &MarketHandler{svc: svc}
}

// Indices GET /market/indices 当日大盘指数列表。
func (h *MarketHandler) Indices(c *gin.Context) {
	indices, err := h.svc.Indices(c.Request.Context())
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, indices)
}

// ListWatchlist GET /market/watchlist 自选股列表。
func (h *MarketHandler) ListWatchlist(c *gin.Context) {
	userID := middleware.GetUserID(c)
	items, err := h.svc.ListWatchlist(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, response.FromWatchlistItems(items))
}

// AddWatchlist POST /market/watchlist 添加自选股。
func (h *MarketHandler) AddWatchlist(c *gin.Context) {
	var req request.AddWatchlistReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	userID := middleware.GetUserID(c)
	item, err := h.svc.AddWatchlist(c.Request.Context(), userID, &req)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, response.FromWatchlistItem(*item))
}

// DeleteWatchlist DELETE /market/watchlist/:id 删除自选股。
func (h *MarketHandler) DeleteWatchlist(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		resp.Fail(c, errcode.ErrInvalidParam.Wrap(err))
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.DeleteWatchlist(c.Request.Context(), userID, id); err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, nil)
}

// WatchlistQuotes GET /market/watchlist/quotes 自选股实时行情。
func (h *MarketHandler) WatchlistQuotes(c *gin.Context) {
	userID := middleware.GetUserID(c)
	quotes, err := h.svc.WatchlistQuotes(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, quotes)
}

// StockDetail GET /market/stocks/:code 个股行情详情。
func (h *MarketHandler) StockDetail(c *gin.Context) {
	code := c.Param("code")
	quote, err := h.svc.StockDetail(c.Request.Context(), code)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, quote)
}

// Kline GET /market/stocks/:code/kline K 线数据。
func (h *MarketHandler) Kline(c *gin.Context) {
	code := c.Param("code")
	period := c.DefaultQuery("period", "day")
	adjust := c.DefaultQuery("adjust", "qfq")
	klines, err := h.svc.Kline(c.Request.Context(), code, period, adjust)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, klines)
}

// Search GET /market/search 股票搜索。
func (h *MarketHandler) Search(c *gin.Context) {
	kw := c.Query("kw")
	if kw == "" {
		resp.Fail(c, errcode.ErrInvalidParam.WithMessage("kw 不能为空"))
		return
	}
	briefs, err := h.svc.Search(c.Request.Context(), kw)
	if err != nil {
		resp.Fail(c, err)
		return
	}
	resp.Success(c, briefs)
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
