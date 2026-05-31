# 守望者股票交易系统 · 后端（warden-backend）

Go 1.22 + Gin + GORM + PostgreSQL + Redis 的后端服务。本仓库为 monorepo 子目录 `backend/`，
严格遵循 [`docs/BACKEND.md`](../docs/BACKEND.md) 的分层规范、统一响应、错误码与 context 传播，
并以 **M1 行情** 作为打通各层的竖切范式。

## 目录结构

```
backend/
├── cmd/server/main.go         # 入口：初始化 DB/Redis(失败降级) → 装配 → 启动 gin
├── internal/
│   ├── handler/               # HTTP 适配（market.go）
│   ├── service/               # 业务编排（market.go，含缓存/降级）
│   ├── repository/            # DB CRUD（watchlist.go / market_quote.go，全程 WithContext）
│   ├── model/                 # GORM 模型（User/WatchlistItem/IndexQuote/StockQuote/Position/Trade）
│   ├── dto/{request,response} # 入参/出参 DTO
│   ├── middleware/            # recovery/logger/cors/ratelimit/timeout/auth
│   ├── integration/market/    # 行情源：IMarketProvider + stub + gotdx(脚手架,build tag)
│   ├── mock/                  # mockgen 生成的 mock
│   └── router/router.go       # 中间件装配(§7.3) + /api 路由
├── pkg/{errcode,response,database,cache}
├── config/{config.go,config.yaml}
├── deploy/{docker-compose.yml,Dockerfile,init.sql}
└── Makefile / .env.example
```

## 环境要求

- **Go 工具链**：模块声明 `go 1.22`（最低版本）。在 **macOS 26 / darwin 25** 上请使用
  **Go ≥ 1.24** 的工具链，否则 1.22 的内部链接器产物缺少 `LC_UUID`，新版 dyld 会
  在运行测试/二进制时报 `missing LC_UUID load command` 而 abort。Go 1.24+ 已修复。
- Docker（用于本地 PostgreSQL/Redis）。

## 快速开始

```bash
cd backend

# 1) 启动基础设施（postgres + redis）
make infra-up        # docker compose -f deploy/docker-compose.yml up -d postgres redis

# 2) 配置（任选其一）
cp .env.example .env # 或直接用 config/config.yaml 默认值

# 3) 编译 & 测试
make build           # go build ./...
make test            # go test ./... -cover

# 4) 启动服务
make run             # go run ./cmd/server  → http://localhost:8080
curl http://localhost:8080/health
curl http://localhost:8080/api/market/indices   # 单用户模式无需 token
```

> 无 PostgreSQL/Redis 时服务仍可启动（降级日志，依赖 DB 的接口会返回错误），
> 行情类接口因默认 stub 数据源可直接返回示例数据。

## M1 行情接口（对齐 [`docs/openapi.yaml`](../docs/openapi.yaml)）

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/market/indices` | 当日大盘指数 |
| GET | `/api/market/watchlist` | 自选股列表 |
| POST | `/api/market/watchlist` | 添加自选 `{stock_code,group_name,remark}` |
| DELETE | `/api/market/watchlist/:id` | 删除自选 |
| GET | `/api/market/watchlist/quotes` | 自选股实时行情（缓存→provider→快照兜底，降级 `stale=true`） |
| GET | `/api/market/stocks/:code` | 个股行情详情 |
| GET | `/api/market/stocks/:code/kline?period=day&adjust=qfq` | K 线 |
| GET | `/api/market/search?kw=` | 股票搜索 |

## 测试（TDD）

- 表驱动单测：`internal/service/market_test.go` 覆盖自选股行情的
  **缓存命中（不外呼）/ provider 调用回写 / provider 失败降级 stale** 三条路径。
- `internal/handler/market_test.go`：Handler 绑定 + userID 注入 + 统一响应。
- `pkg/response`、`pkg/errcode`：统一响应与错误码基础单测。
- Mock 用 `go.uber.org/mock`（mockgen 源码模式生成于 `internal/mock`）：

```bash
make mock   # 接口变更后重新生成
```

- 所有外部依赖（行情源/DB/缓存）在单测中以 mock 替换，**禁止真实外呼**。
- CI：`make test-race`（`go test ./... -race -cover`，竞态检测走 cgo 外部链接）。

## 配置

`config/config.go` 用 viper 读取 `config/config.yaml`，并以 BACKEND.md §7.2 的扁平
环境变量覆盖（`APP_PORT`/`SINGLE_USER_MODE`/`JWT_SECRET`/`PG_*`/`REDIS_*`/`MARKET_*`/`CONFIG_ENC_KEY`）。

## 待办（TODO）

- [ ] **接入真实 gotdx 行情源**：当前 `github.com/bensema/gotdx` 最新版要求 `go >= 1.26`，
      与本模块锚定的 `go 1.22` 不兼容，故默认行情源为 `stubProvider`（返回示例数据）。
      真实实现脚手架已按 BACKEND.md §5 落地于 `internal/integration/market/gotdx_pool.go`、
      `gotdx_provider.go`、`gotdx_mapper.go`，均带 **构建标签 `//go:build gotdx`**（默认构建排除）。
      接入步骤：
      1. 将 `go.mod` 的 Go 版本提升至 gotdx 要求的版本并 `go get github.com/bensema/gotdx`；
      2. 按真实 gotdx API 调整 `gotdx_mapper.go` 中的字段映射与 `applyAdjust` 复权计算
         （以「能编译」为准，用适配函数封装差异，保持 `IMarketProvider` 接口不变）；
      3. `MARKET_PROVIDER=gotdx` 并以 `go build -tags gotdx ./...` 构建。
      Service/Handler 无需改动（仅依赖 `IMarketProvider` 接口）。
- [ ] M2~M7 各模块 handler/service/repository（按本竖切范式扩展）。
- [ ] `pkg/crypto` AES-GCM、`pkg/mq` RabbitMQ、`scheduler` cron 调度器。
- [ ] 集成测试（`test/`，dockertest 起测试库跑真实 SQL）。
