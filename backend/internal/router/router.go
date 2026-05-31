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

// Handlers 聚合各业务 Handler，便于注入与扩展（后续 M3~M7 在此追加）。
type Handlers struct {
	Market   *handler.MarketHandler
	Strategy *handler.StrategyHandler
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
	r.Use(middleware.Timeout(
		time.Duration(cfg.Timeout.Seconds)*time.Second,
		timeoutOverrides(cfg),
	))

	// 健康检查（不鉴权）。
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	api.Use(middleware.Auth(cfg.JWT.Secret, cfg.App.SingleUserMode))

	registerMarketRoutes(api, h.Market)
	registerStrategyRoutes(api, h.Strategy)
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

// timeoutOverrides 按 gin 路由模式声明慢接口的超时覆盖，
// 由 middleware.Timeout 按最长前缀匹配生效。
// 粗筛/预览需批量拉行情，远超普通 CRUD 的 10s 默认值。
func timeoutOverrides(cfg *config.Config) map[string]time.Duration {
	screenTimeout := time.Duration(cfg.Screen.TimeoutSeconds) * time.Second
	if screenTimeout <= 0 {
		return nil
	}
	return map[string]time.Duration{
		"/api/strategies/screen/preview": screenTimeout,
		"/api/strategies/:id/screen":     screenTimeout, // 覆盖 :id/screen 及 :id/screen/*
	}
}

func registerStrategyRoutes(api *gin.RouterGroup, h *handler.StrategyHandler) {
	if h == nil {
		return
	}
	g := api.Group("/strategies")
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		// 静态路径需在 :id 之前声明，避免与参数路由冲突。
		g.GET("/indicators/catalog", h.Catalog)
		g.GET("/templates", h.Templates)
		g.POST("/screen/preview", h.PreviewScreen)

		g.GET("/:id", h.Get)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
		g.POST("/:id/copy", h.Copy)
		g.PUT("/:id/indicators", h.UpdateIndicators)
		g.GET("/:id/skill", h.GetSkill)
		g.PUT("/:id/skill", h.SaveSkill)
		g.POST("/:id/screen", h.RunScreen)
		g.GET("/:id/screen/latest", h.ScreenLatest)
		g.GET("/:id/screen/:taskId", h.ScreenResult)
	}
}
