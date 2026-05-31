// Command server 是守望者后端 HTTP 服务入口。
//
// 设计原则：DB/Redis 初始化失败时仅记录日志并降级（不 panic），
// 方便本地无依赖时也能 go build / 启动（见任务要求）。
package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"warden/config"
	"warden/internal/handler"
	"warden/internal/integration/market"
	"warden/internal/model"
	"warden/internal/repository"
	"warden/internal/router"
	"warden/internal/service"
	"warden/pkg/cache"
	"warden/pkg/database"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		return
	}

	db := initDB(cfg)
	redisCli := initRedis(cfg)

	// 行情数据源：默认 stub（gotdx 需 -tags gotdx 且 Go>=1.26，见 README 待办）。
	provider := market.NewProvider(cfg.Market.Provider, cfg.Market.GotdxMaxConn)

	watchRepo := repository.NewWatchlistRepository(db)
	quoteRepo := repository.NewMarketQuoteRepository(db)
	strategyRepo := repository.NewStrategyRepository(db)
	screenRepo := repository.NewScreenResultRepository(db)

	var quoteCache service.QuoteCache
	if redisCli != nil {
		quoteCache = service.NewRedisQuoteCache(redisCli, 5*time.Second)
	}

	marketSvc := service.NewMarketService(provider, watchRepo, quoteRepo, quoteCache)
	marketHandler := handler.NewMarketHandler(marketSvc)

	strategySvc := service.NewStrategyService(
		strategyRepo, screenRepo, watchRepo, provider,
		service.WithScreenConcurrency(cfg.Screen.Concurrency),
	)
	strategyHandler := handler.NewStrategyHandler(strategySvc)

	r := router.New(cfg, router.Handlers{Market: marketHandler, Strategy: strategyHandler})

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	slog.Info("守望者后端启动", "addr", addr, "single_user_mode", cfg.App.SingleUserMode, "market_provider", cfg.Market.Provider)
	if err := r.Run(addr); err != nil {
		slog.Error("HTTP 服务退出", "error", err)
	}
}

func initDB(cfg *config.Config) *gorm.DB {
	db, err := database.New(database.Config{
		Host:         cfg.Postgres.Host,
		Port:         cfg.Postgres.Port,
		User:         cfg.Postgres.User,
		Password:     cfg.Postgres.Password,
		DBName:       cfg.Postgres.DB,
		SSLMode:      cfg.Postgres.SSLMode,
		MaxOpenConns: cfg.Postgres.MaxOpenConns,
		MaxIdleConns: cfg.Postgres.MaxIdleConns,
	})
	if err != nil {
		slog.Warn("PostgreSQL 连接失败，降级运行（依赖 DB 的接口将不可用）", "error", err)
		return nil
	}
	// 开发期自动迁移核心表（生产以 deploy/init.sql 为准）。
	// 注意：必须逐 model 调用，单个 model 的索引/约束差异不应阻断后续表的创建，
	//（例如 init.sql 中 users.username 使用匿名 UNIQUE 时，GORM 想 DROP "uni_users_username" 会失败，
	// 一次性传入所有 model 会导致后面的 strategy_screen_results 等表静默漏建）。
	models := []any{
		&model.User{}, &model.WatchlistItem{},
		&model.IndexQuote{}, &model.StockQuote{},
		&model.Position{}, &model.Trade{},
		&model.Strategy{}, &model.StrategyIndicator{},
		&model.StrategySkill{}, &model.StrategyScreenResult{},
	}
	for _, m := range models {
		if mErr := db.AutoMigrate(m); mErr != nil {
			slog.Warn("AutoMigrate 单表失败（已跳过，继续后续表）", "model", fmt.Sprintf("%T", m), "error", mErr)
		}
	}
	slog.Info("PostgreSQL 已连接")
	return db
}

func initRedis(cfg *config.Config) *redis.Client {
	cli, err := cache.New(cache.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		slog.Warn("Redis 连接失败，降级运行（行情缓存关闭）", "error", err)
		return nil
	}
	slog.Info("Redis 已连接")
	return cli
}
