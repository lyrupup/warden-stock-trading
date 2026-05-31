// Package router 装配中间件与路由（中间件顺序见 BACKEND.md §7.3）。
package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"warden/config"
	"warden/internal/handler"
	"warden/internal/middleware"
)

// Handlers 聚合各业务 Handler，便于注入与扩展（后续 M2~M7 在此追加）。
type Handlers struct {
	Market *handler.MarketHandler
}

// New 创建 gin 引擎，按规范顺序装配中间件并挂载 /api 路由组。
func New(cfg *config.Config, h Handlers) *gin.Engine {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// 中间件装配顺序：Recovery → Logger → CORS → RateLimit → Timeout(context) → Auth。
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(cfg.RateLimit.RPS, cfg.RateLimit.Burst))
	r.Use(middleware.Timeout(time.Duration(cfg.Timeout.Seconds) * time.Second))

	// 健康检查（不鉴权）。
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	api.Use(middleware.Auth(cfg.JWT.Secret, cfg.App.SingleUserMode))

	registerMarketRoutes(api, h.Market)
	return r
}

func registerMarketRoutes(api *gin.RouterGroup, h *handler.MarketHandler) {
	if h == nil {
		return
	}
	m := api.Group("/market")
	{
		m.GET("/indices", h.Indices)
		m.GET("/watchlist", h.ListWatchlist)
		m.POST("/watchlist", h.AddWatchlist)
		m.DELETE("/watchlist/:id", h.DeleteWatchlist)
		m.GET("/watchlist/quotes", h.WatchlistQuotes)
		m.GET("/stocks/:code", h.StockDetail)
		m.GET("/stocks/:code/kline", h.Kline)
		m.GET("/search", h.Search)
	}
}
