# 守望者股票交易系统 · 后端（warden-backend）

Go 1.26 + Gin + GORM + PostgreSQL + Redis 的后端服务。本仓库为 monorepo 子目录 `backend/`，
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
│   ├── integration/market/    # 行情源：IMarketProvider + stub(默认) + gotdx 真实实现(build tag)
│   ├── mock/                  # mockgen 生成的 mock
│   └── router/router.go       # 中间件装配(§7.3) + /api 路由
├── pkg/{errcode,response,database,cache}
├── config/{config.go,config.yaml}
├── deploy/{docker-compose.yml,Dockerfile,init.sql}
└── Makefile / .env.example
```

## 环境要求

- **Go 工具链**：模块声明 `go 1.26.0`（`github.com/bensema/gotdx` 要求 Go ≥ 1.26）。
  本机若为 Go 1.24/1.25，设 `GOTOOLCHAIN=auto`（默认值）即可在首次构建时自动拉取
  并切换到 go1.26 工具链，无需手动安装。
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

# 4) 启动服务（二选一）
make run             # go run ./cmd/server            → stub 示例行情（无需外网）
make run-gotdx       # go run -tags gotdx ./cmd/server → 真实通达信实时行情
curl http://localhost:8080/health
curl http://localhost:8080/api/market/indices   # 单用户模式无需 token
```

> 无 PostgreSQL/Redis 时服务仍可启动（降级日志，依赖 DB 的接口会返回错误）。
>
> **行情数据源**：默认构建（`make run` / `make build`）行情源回退 `stubProvider`，
> 返回确定性示例数据，便于离线开发与单测。要拿**真实通达信行情**，用
> `make run-gotdx`（即 `-tags gotdx`，编译进 `//go:build gotdx` 文件直连通达信主站），
> 并确保 `market.provider=gotdx`（默认）或 `MARKET_PROVIDER=gotdx`。

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

- [x] **接入真实 gotdx 行情源**：`go.mod` 已锚定 `go 1.26.0` 并 `go get github.com/bensema/gotdx`。
      真实实现位于 `internal/integration/market/`（均带 **构建标签 `//go:build gotdx`**）：
      - `gotdx_pool.go`：通达信连接池（单连接非并发安全，借出独占/用完归还/坏连接丢弃），
        内置主站地址池 + 连接前测速优选最快节点；
      - `gotdx_provider.go`：`Indices`(`StockIndexInfo`)、`Quotes`(`StockQuotesDetail`)、
        `Kline`(`StockKLine`，**服务端复权** qfq/hfq)、`Search`（懒加载全市场证券名索引本地过滤，
        启动时后台预热）；所有外呼包 panic 兜底，畸形/空响应转 error 优雅降级；
      - `gotdx_mapper.go`：`proto.*` → warden model 字段映射、市场代码/K线周期/复权枚举转换。
      用 `make run-gotdx`（或 `go build -tags gotdx ./...` + `MARKET_PROVIDER=gotdx`）启用；
      Service/Handler 无改动（仅依赖 `IMarketProvider` 接口）。
      > 注：单次 K 线条数上限 480（通达信单包上限不足 800）；换手率 = gotdx `Turnover`÷10000
      > （其量纲为「真实换手率% × 10000」，流通股单位为万股，故还原为百分比）。
- [ ] M2~M7 各模块 handler/service/repository（按本竖切范式扩展）。
- [ ] `pkg/crypto` AES-GCM、`pkg/mq` RabbitMQ、`scheduler` cron 调度器。
- [ ] 集成测试（`test/`，dockertest 起测试库跑真实 SQL）。
