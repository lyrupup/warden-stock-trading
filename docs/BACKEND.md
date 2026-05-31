# 守望者股票交易系统 · 后端技术开发文档

> Warden Stock Trading System · Backend Technical Design
>
> 文档版本：v1.0　|　创建日期：2026-05-30　|　依据：[`docs/PRD.md`](./PRD.md)
>
> 本文档包含：技术栈、目录结构、**数据库设计**、**API 接口文档**、**TDD 测试文档**、核心功能模块设计方案。与《前端技术开发文档》并行开发，通过统一响应契约与本文 §4 API 文档对齐。

---

## 1. 技术栈与架构

### 1.1 技术栈

| 分类 | 选型 | 说明 |
|------|------|------|
| 语言/框架 | **Go 1.22 + Gin** | 高并发、强类型核心服务 |
| ORM | **GORM** | PostgreSQL 驱动，所有操作传播 context |
| 数据库 | **PostgreSQL 16** | 业务主存储 |
| 缓存 | **Redis 7** | 行情快照缓存、限流、会话、分布式锁 |
| 消息/异步 | **RabbitMQ 3**（可选） | 回测、AI 分析、推送等耗时任务异步化 |
| 调度 | **robfig/cron v3** | M5 定时任务调度 |
| 行情数据源 | **bensema/gotdx**（直连通达信，主力）+ Tushare HTTP（补充） | 见 §5 M1 选型，经连接池封装接入 |
| 鉴权 | **JWT**（golang-jwt） | Bearer Token，多用户预留 |
| 配置 | **viper** + 环境变量 | 不硬编码 |
| 校验 | **go-playground/validator** | DTO 参数校验 |
| 测试 | **testing + testify + go.uber.org/mock** | TDD、表驱动、Mock |
| 文档 | **swaggo/swag**（OpenAPI） | 生成 API 文档供前端对齐 |
| 加密 | **AES-GCM** | 用户敏感配置（API Key）加密存储 |

### 1.2 分层架构

```
请求 → Router → Middleware → Handler → Service → Repository → PostgreSQL
                                          ↓             ↓
                                     Cache(Redis)   外部集成层
                                          ↓        (行情源/大模型/飞书)
```

| 层 | 职责 | 禁止 |
|----|------|------|
| Handler | 参数绑定/校验、调 Service、组装响应 | 业务逻辑、直连 DB |
| Service | 业务编排、事务、调 Repository/外部集成 | 处理 HTTP |
| Repository | DB CRUD（`WithContext`） | 业务逻辑 |
| Integration | 行情源/大模型/飞书 client | 业务逻辑 |

### 1.3 目录结构

```
backend/
├── cmd/server/main.go
├── internal/
│   ├── handler/            # market.go strategy.go position.go risk.go task.go ai.go settings.go auth.go
│   ├── service/            # 同名业务逻辑层
│   ├── repository/         # 数据访问层
│   ├── model/              # GORM 模型
│   ├── dto/                # request/ response/
│   ├── middleware/         # auth ratelimit timeout logger cors recovery
│   ├── integration/        # market/(行情源) llm/(大模型) feishu/(飞书)
│   ├── scheduler/          # cron 调度器 + 任务执行器
│   ├── mock/               # mockgen 生成
│   └── router/router.go
├── pkg/
│   ├── errcode/            # 统一错误码
│   ├── response/           # 统一响应
│   ├── database/           # PG 连接
│   ├── cache/              # Redis 封装
│   ├── mq/                 # RabbitMQ 封装
│   ├── crypto/             # AES 加解密
│   └── utils/
├── config/config.yaml
├── deploy/{docker-compose.yml,Dockerfile,init.sql}
├── test/                   # 集成测试
└── go.mod
```

### 1.4 统一响应与错误码

```go
// pkg/response
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
// Success: code=0；分页 data={list,total,page,size}
```

错误码段位：`10xxx` 通用、`20xxx` 行情、`21xxx` 策略、`22xxx` 持仓、`23xxx` 风控、`24xxx` 任务、`25xxx` AI、`26xxx` 配置、`40xxx` 鉴权。

---

## 2. 多用户与鉴权设计

> PRD 要求：V1 可单用户运行，但数据模型与接口必须为多用户预留。

- **所有业务表含 `user_id`**，Repository 查询强制带 `user_id` 条件，杜绝越权。
- **鉴权中间件 `AuthMiddleware`**：解析 JWT → 注入 `user_id` 到 `gin.Context`。
- **单用户模式开关**：配置 `single_user_mode: true` 时，`AuthMiddleware` 在无 token 时自动注入默认用户（`user_id=1`），保证 V1 可直接使用；多用户上线时关掉开关即可，无需改业务代码。
- Service/Repository 一律从 context 取 `userID`，不信任前端传入。

```go
func GetUserID(c *gin.Context) uint { v, _ := c.Get("user_id"); return v.(uint) }
```

---

## 3. 数据库设计文档

### 3.1 设计约定

- 表名：小写 + 下划线 + 复数；字段：小写下划线；软删除 `deleted_at`。
- 公共基类：`id BIGSERIAL`、`created_at`、`updated_at`、`deleted_at`。
- 业务表统一含 `user_id BIGINT NOT NULL`，并建 `idx_<table>_user_id`。
- 金额/价格用 `NUMERIC(20,4)`，比率用 `NUMERIC(10,4)`，避免浮点误差。
- 敏感字段（API Key）密文存储（AES-GCM），表中存 `*_cipher`。

### 3.2 ER 概览

```
users 1──N watchlist_items
users 1──N strategies 1──N strategy_indicators
                     1──1 strategy_skills (含版本)
                     1──N backtest_results
users 1──N positions 1──N trades
users 1──N position_analysis_reports
users 1──N premarket_plans
users 1──N alert_records
users 1──N scheduled_tasks 1──N task_execution_logs
users 1──N analysis_conversations 1──N conversation_messages
users 1──N stock_analysis_reports        (按 stock_code 归档)
users 1──N user_configs                  (分组键值, 敏感加密)
公共:  index_quotes / stock_quotes (行情快照, 不属单用户)
```

### 3.3 核心表结构（init.sql 摘要）

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 用户（多用户预留）
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(128) NOT NULL DEFAULT '',
    nickname VARCHAR(64) NOT NULL DEFAULT '',
    avatar VARCHAR(256) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
-- 预置默认用户（单用户模式）
INSERT INTO users (id, username, nickname) VALUES (1, 'default', '守望者')
    ON CONFLICT DO NOTHING;

-- M1 自选股
CREATE TABLE IF NOT EXISTS watchlist_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    stock_code VARCHAR(16) NOT NULL,
    stock_name VARCHAR(64) NOT NULL DEFAULT '',
    group_name VARCHAR(64) NOT NULL DEFAULT 'default',
    remark VARCHAR(256) NOT NULL DEFAULT '',
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (user_id, stock_code, deleted_at)
);
CREATE INDEX IF NOT EXISTS idx_watchlist_items_user_id ON watchlist_items(user_id);

-- M1 行情快照（公共，缓存兜底）
CREATE TABLE IF NOT EXISTS index_quotes (
    id BIGSERIAL PRIMARY KEY,
    index_code VARCHAR(16) NOT NULL,
    index_name VARCHAR(64) NOT NULL,
    price NUMERIC(20,4), change_amount NUMERIC(20,4), change_percent NUMERIC(10,4),
    volume NUMERIC(20,0), amount NUMERIC(24,4),
    trade_date DATE NOT NULL,
    snapshot_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_index_quotes_code_date ON index_quotes(index_code, trade_date);

CREATE TABLE IF NOT EXISTS stock_quotes (
    id BIGSERIAL PRIMARY KEY,
    stock_code VARCHAR(16) NOT NULL,
    stock_name VARCHAR(64) NOT NULL DEFAULT '',
    price NUMERIC(20,4), open NUMERIC(20,4), high NUMERIC(20,4), low NUMERIC(20,4), prev_close NUMERIC(20,4),
    change_percent NUMERIC(10,4), volume NUMERIC(20,0), amount NUMERIC(24,4),
    turnover_rate NUMERIC(10,4),
    trade_date DATE NOT NULL,
    snapshot_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_stock_quotes_code_date ON stock_quotes(stock_code, trade_date);

-- M2 策略
CREATE TABLE IF NOT EXISTS strategies (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    tags VARCHAR(256) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_strategies_user_id ON strategies(user_id);

-- 指标定义（结构化条件存 JSONB）
CREATE TABLE IF NOT EXISTS strategy_indicators (
    id BIGSERIAL PRIMARY KEY,
    strategy_id BIGINT NOT NULL,
    conditions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_strategy_indicators_strategy_id ON strategy_indicators(strategy_id);

-- 策略 skill.md（含版本）
CREATE TABLE IF NOT EXISTS strategy_skills (
    id BIGSERIAL PRIMARY KEY,
    strategy_id BIGINT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_strategy_skills_strategy_id ON strategy_skills(strategy_id);

-- 回测结果
CREATE TABLE IF NOT EXISTS backtest_results (
    id BIGSERIAL PRIMARY KEY,
    strategy_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    params JSONB NOT NULL DEFAULT '{}',
    status SMALLINT NOT NULL DEFAULT 0,    -- 0 pending 1 running 2 done 3 failed
    metrics JSONB,                          -- 收益率/最大回撤/胜率/夏普
    curve JSONB,                            -- 净值曲线
    error_msg VARCHAR(512) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_backtest_results_strategy_id ON backtest_results(strategy_id);

-- M3 持仓
CREATE TABLE IF NOT EXISTS positions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    stock_code VARCHAR(16) NOT NULL,
    stock_name VARCHAR(64) NOT NULL DEFAULT '',
    quantity NUMERIC(20,4) NOT NULL DEFAULT 0,      -- 持仓数量
    avg_cost NUMERIC(20,4) NOT NULL DEFAULT 0,      -- 移动加权成本均价
    total_cost NUMERIC(20,4) NOT NULL DEFAULT 0,    -- 当前持仓成本
    realized_pnl NUMERIC(20,4) NOT NULL DEFAULT 0,  -- 已实现盈亏累计
    status SMALLINT NOT NULL DEFAULT 1,             -- 1 持有 2 已清仓
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_positions_user_id ON positions(user_id);

-- 交易记录
CREATE TABLE IF NOT EXISTS trades (
    id BIGSERIAL PRIMARY KEY,
    position_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    side SMALLINT NOT NULL,                          -- 1 买入 2 卖出
    price NUMERIC(20,4) NOT NULL,
    quantity NUMERIC(20,4) NOT NULL,
    fee NUMERIC(20,4) NOT NULL DEFAULT 0,            -- 手续费
    tax NUMERIC(20,4) NOT NULL DEFAULT 0,            -- 印花税等
    realized_pnl NUMERIC(20,4) NOT NULL DEFAULT 0,   -- 卖出时计算的本笔已实现盈亏
    remark VARCHAR(256) NOT NULL DEFAULT '',
    traded_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_trades_position_id ON trades(position_id);
CREATE INDEX IF NOT EXISTS idx_trades_user_id ON trades(user_id);

-- M4 持仓分析报告 / 盘前预案 / 告警
CREATE TABLE IF NOT EXISTS position_analysis_reports (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    report_date DATE NOT NULL,
    content TEXT NOT NULL DEFAULT '',     -- Markdown
    risk_items JSONB,                     -- [{level,target,desc,advice}]
    risk_level SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_par_user_date ON position_analysis_reports(user_id, report_date);

CREATE TABLE IF NOT EXISTS premarket_plans (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    trade_date DATE NOT NULL,
    plans JSONB NOT NULL DEFAULT '[]',    -- 按个股的应对策略
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_premarket_user_date ON premarket_plans(user_id, trade_date);

CREATE TABLE IF NOT EXISTS alert_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    stock_code VARCHAR(16) NOT NULL DEFAULT '',
    rule VARCHAR(128) NOT NULL DEFAULT '',
    level SMALLINT NOT NULL DEFAULT 1,
    message VARCHAR(512) NOT NULL DEFAULT '',
    pushed SMALLINT NOT NULL DEFAULT 0,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alert_records_user_id ON alert_records(user_id);

-- M5 定时任务
CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    task_type VARCHAR(32) NOT NULL,        -- fetch_data/analysis/report/push
    cron_expr VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',   -- 类型相关参数（含推送目标）
    trading_day_only BOOLEAN NOT NULL DEFAULT TRUE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_user_id ON scheduled_tasks(user_id);

CREATE TABLE IF NOT EXISTS task_execution_logs (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,     -- 0 running 1 success 2 failed
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    output TEXT NOT NULL DEFAULT '',
    error_msg VARCHAR(512) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_task_logs_task_id ON task_execution_logs(task_id);

-- M6 AI 会话与报告
CREATE TABLE IF NOT EXISTS analysis_conversations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    strategy_id BIGINT NOT NULL,
    stock_code VARCHAR(16) NOT NULL,
    title VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_conv_user_code ON analysis_conversations(user_id, stock_code);

CREATE TABLE IF NOT EXISTS conversation_messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    role VARCHAR(16) NOT NULL,              -- user/assistant/system
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_conv_msg_conv_id ON conversation_messages(conversation_id);

-- 分析报告（按股票代码归档）
CREATE TABLE IF NOT EXISTS stock_analysis_reports (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    stock_code VARCHAR(16) NOT NULL,
    strategy_id BIGINT NOT NULL,
    conversation_id BIGINT,
    title VARCHAR(128) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',       -- 结构化报告 Markdown
    tags VARCHAR(256) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sar_user_code ON stock_analysis_reports(user_id, stock_code);

-- M7 用户配置（分组键值，敏感加密）
CREATE TABLE IF NOT EXISTS user_configs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    config_group VARCHAR(32) NOT NULL,      -- ai/feishu/datasource/preference/profile
    config_key VARCHAR(64) NOT NULL,
    config_value TEXT NOT NULL DEFAULT '',  -- 非敏感明文
    value_cipher BYTEA,                     -- 敏感字段密文(AES-GCM)
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, config_group, config_key)
);
CREATE INDEX IF NOT EXISTS idx_user_configs_user_id ON user_configs(user_id);
```

### 3.4 GORM 模型示例

```go
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Position struct {
    BaseModel
    UserID      uint            `gorm:"index;not null" json:"user_id"`
    StockCode   string          `gorm:"size:16;not null" json:"stock_code"`
    StockName   string          `gorm:"size:64" json:"stock_name"`
    Quantity    decimal.Decimal `gorm:"type:numeric(20,4)" json:"quantity"`
    AvgCost     decimal.Decimal `gorm:"type:numeric(20,4)" json:"avg_cost"`
    TotalCost   decimal.Decimal `gorm:"type:numeric(20,4)" json:"total_cost"`
    RealizedPnL decimal.Decimal `gorm:"type:numeric(20,4)" json:"realized_pnl"`
    Status      int8            `gorm:"default:1" json:"status"`
}

func (Position) TableName() string { return "positions" }
```

> 金额统一用 `shopspring/decimal`，避免浮点误差。

---

## 4. API 接口文档

> 📄 机器可读的接口契约见 [`docs/openapi.yaml`](./openapi.yaml)（OpenAPI 3.0）。前端可用其生成类型/Mock，后端可用 swaggo 同步维护。本节为人读摘要，二者保持一致。

### 4.1 通用约定

- BasePath：`/api`
- 鉴权：`Authorization: Bearer <token>`（单用户模式可省略）。
- 响应体：`{ "code": 0, "message": "ok", "data": ... }`。
- 分页：query `?page=1&size=20`，响应 `data={list,total,page,size}`。
- 时间：ISO8601。
- **数值精度（重要约定）**：所有**金额 / 价格 / 数量 / 比率**字段后端统一用 `shopspring/decimal.Decimal`（DB 为 `numeric`）以保证精度，其默认 `MarshalJSON` **序列化为带引号的 JSON 字符串**（如 `"10.5000"`、`"1.94"`）。
  - 即：行情（price/open/high/low/change_percent/volume/amount/turnover_rate）、持仓与交易（quantity/avg_cost/total_cost/pnl/fee/tax）、回测指标、账户总览等数值字段，**线上类型为 string（decimal）**，对应 openapi 的 `Decimal` schema（`type: string`）。
  - 入参：DTO 中的 decimal 字段后端可同时解析数字与字符串（`10.5` 或 `"10.5"`）。
  - 前端约定：必须经统一 decimal 工具（`lib/decimal.ts` 的 `toNumber` / `coerceDecimalFields`）转 number 后再计算或格式化，**禁止直接对其做算术或 `.toFixed()`**。
  - 取舍说明：保精度优先（金融场景）；如需改为输出 JSON 数字，可全局设 `decimal.MarshalJSONWithoutQuotes = true`，但需评估大数精度，且要同步更新 openapi 与前端。当前默认采用 **decimal 字符串** 方案。

### 4.2 鉴权 Auth

| Method | Path | 说明 | Body/Query |
|--------|------|------|-----------|
| POST | `/auth/login` | 登录获取 token | `{username,password}` |
| POST | `/auth/register` | 注册（多用户预留） | `{username,password,nickname}` |
| GET | `/auth/me` | 当前用户信息 | - |

### 4.3 M1 行情 Market

| Method | Path | 说明 |
|--------|------|------|
| GET | `/market/indices` | 当日大盘指数列表 |
| GET | `/market/overview` | 市场概览（涨跌家数等） |
| GET | `/market/watchlist` | 自选股列表 |
| POST | `/market/watchlist` | 添加自选 `{stock_code,group_name,remark}` |
| DELETE | `/market/watchlist/:id` | 删除自选 |
| GET | `/market/watchlist/quotes` | 自选股实时行情 |
| GET | `/market/stocks/:code` | 个股行情详情 |
| GET | `/market/stocks/:code/kline?period=day` | K 线数据 |
| GET | `/market/search?kw=` | 股票搜索 |

### 4.4 M2 策略 Strategy

| Method | Path | 说明 |
|--------|------|------|
| GET | `/strategies` | 策略列表（支持 `?kw=&tag=`） |
| POST | `/strategies` | 新建策略 |
| GET | `/strategies/:id` | 策略详情（含指标、skill） |
| PUT | `/strategies/:id` | 更新策略 |
| DELETE | `/strategies/:id` | 删除策略 |
| POST | `/strategies/:id/copy` | 复制策略 |
| PUT | `/strategies/:id/indicators` | 更新指标定义（JSON 条件） |
| GET | `/strategies/:id/skill` | 获取 skill.md |
| PUT | `/strategies/:id/skill` | 保存 skill.md（生成新版本） |
| POST | `/strategies/:id/backtest` | 发起回测 → `{taskId}` |
| GET | `/strategies/:id/backtest/:taskId` | 查询回测结果 |
| POST | `/strategies/:id/screen` | 运行选股（可选） |

请求示例（新建策略）：

```json
POST /api/strategies
{
  "name": "低估值蓝筹",
  "description": "PE<15 且 ROE>15% 的大盘价值股",
  "tags": "价值,蓝筹",
  "indicators": { "logic": "and", "rules": [
    {"indicator":"pe","op":"<","value":15},
    {"indicator":"roe","op":">","value":15}
  ]}
}
```

### 4.5 M3 持仓 Position

| Method | Path | 说明 |
|--------|------|------|
| GET | `/positions` | 持仓列表（含浮盈，行情联动） |
| POST | `/positions` | 新建持仓 |
| GET | `/positions/:id` | 持仓详情 |
| PUT | `/positions/:id` | 编辑持仓 |
| DELETE | `/positions/:id` | 删除持仓 |
| GET | `/positions/summary` | 账户总览 |
| GET | `/positions/:id/trades` | 交易流水（分页） |
| POST | `/positions/:id/trades` | 录入交易（自动重算均价/盈亏） |
| DELETE | `/positions/:id/trades/:tid` | 删除交易记录（重算） |

录入交易请求：

```json
POST /api/positions/12/trades
{ "side": 1, "price": 10.50, "quantity": 1000, "fee": 5, "tax": 0, "traded_at": "2026-05-30T09:35:00+08:00", "remark": "建仓" }
```

### 4.6 M4 风控 Risk

| Method | Path | 说明 |
|--------|------|------|
| POST | `/risk/analysis` | 触发每日持仓 AI 分析（流式 SSE，见 §4.10） |
| GET | `/risk/analysis?date=` | 查询某日分析报告 |
| GET | `/risk/analysis/history` | 历史分析列表（分页） |
| POST | `/risk/premarket` | 生成明日盘前预案 |
| GET | `/risk/premarket?date=` | 查询盘前预案 |
| GET | `/risk/alerts` | 告警记录列表 |

### 4.7 M5 定时任务 Task

| Method | Path | 说明 |
|--------|------|------|
| GET | `/tasks` | 任务列表 |
| POST | `/tasks` | 新建任务 |
| GET | `/tasks/:id` | 任务详情 |
| PUT | `/tasks/:id` | 编辑任务 |
| DELETE | `/tasks/:id` | 删除任务 |
| PATCH | `/tasks/:id/toggle` | 启用/停用 |
| POST | `/tasks/:id/run` | 立即执行一次 |
| GET | `/tasks/:id/logs` | 执行记录（分页） |

新建任务示例（盘后推送报告到飞书）：

```json
POST /api/tasks
{
  "name": "盘后持仓报告推送",
  "task_type": "push",
  "cron_expr": "0 30 15 * * 1-5",
  "trading_day_only": true,
  "payload": { "source": "position_analysis", "channel": "feishu_bot" }
}
```

### 4.8 M6 AI 分析 AI

| Method | Path | 说明 |
|--------|------|------|
| GET | `/ai/conversations` | 会话列表 |
| POST | `/ai/conversations` | 新建会话 `{strategy_id,stock_code}` |
| GET | `/ai/conversations/:id/messages` | 会话消息 |
| POST | `/ai/stream` | 发送消息（SSE 流式返回，见 §4.10） |
| POST | `/ai/conversations/:id/report` | 由对话生成结构化报告 |
| GET | `/ai/reports?stock_code=&strategy_id=&date=` | 报告库检索 |
| GET | `/ai/reports/grouped` | 按股票代码分组的报告库 |
| GET | `/ai/reports/:id` | 报告详情 |
| POST | `/ai/reports/:id/push` | 推送报告到飞书 |

### 4.9 M7 配置 Settings

| Method | Path | 说明 |
|--------|------|------|
| GET | `/settings` | 获取全部配置（敏感字段脱敏） |
| PUT | `/settings/:group` | 按组保存配置（ai/feishu/datasource/preference/profile） |
| POST | `/settings/feishu/test` | 测试飞书机器人连通性 |
| POST | `/settings/ai/test` | 测试大模型连通性 |

### 4.10 SSE 流式接口约定（M4 / M6）

```
POST /api/ai/stream    (Accept: text/event-stream)
event: message\ndata: {"delta":"分析中..."}\n\n
event: message\ndata: {"delta":"该股..."}\n\n
event: done\ndata: {"reportId": 123}\n\n
event: error\ndata: {"code":25001,"message":"llm error"}\n\n
```

---

## 5. 核心功能模块设计方案

### M1 行情模块设计

- **数据流**：Service 优先读 Redis 缓存（`warden:market:quote:{code}`，TTL 随交易时段调整）→ 未命中调 `integration/market` 拉取 → 回写缓存 + `stock_quotes` 快照表兜底。
- **集成抽象**：`IMarketProvider` 接口，便于切换数据源（实现注入）。
- **降级**：外部源失败时返回最近快照并在响应 `data.stale=true` 标记。

```go
type IMarketProvider interface {
    Indices(ctx context.Context) ([]model.IndexQuote, error)
    Quotes(ctx context.Context, codes []string) ([]model.StockQuote, error)
    Kline(ctx context.Context, code, period, adjust string) ([]model.Kline, error)
    Search(ctx context.Context, kw string) ([]model.StockBrief, error)
}
```

#### 行情数据源 / SDK 选型

> A 股**没有官方 Go SDK**，主流是封装新浪 / 腾讯 / 东方财富 / 通达信的公开接口，或对接 Tushare / BaoStock。本系统通过 `IMarketProvider` 接口屏蔽底层差异，可按需切换或组合多源。
>
> **选型结论：以 `bensema/gotdx`（直连通达信协议）为主力行情源。** 经评估，社区里多数 A 股 Go 封装库（efinance-go / akshare-go / baostock-go 等）star 偏低、维护者单一、存在停更风险，不宜作为核心强依赖；而 `gotdx` 在活跃度、能力覆盖、健壮性上明显更优，作为主源更稳妥。

**主力源：`bensema/gotdx`**（189★/85fork，MIT，持续维护，纯 Go 直连通达信 TCP 协议）

| 维度 | 说明 |
|------|------|
| 优势 | 无需 token / 无需 HTTP 抓取；内置 host 地址池 + **TCP 测速选最快节点** + 超时/重试可配；覆盖快照、K线、分时、逐笔、F10、财务、除权除息、板块、资金流、集合竞价、异动 |
| 覆盖需求 | M1 实时行情/K线/分时/指数 + M2 回测历史K线（+复权） + M2 基本面（F10/财务） + M4 风控（资金流/异动/板块） 一站式满足 |
| 注意点 | ① 通达信为**非官方逆向协议**，服务端可能变更/限频（库的地址池+测速+重试已缓解）；② TDX **单连接非并发安全**，后端必须封装**连接池**借还；③ 实时性为**快照级（约 3–5s）**，非毫秒 tick，个人系统足够；④ K线默认**不复权**，需用 `GetXDXRInfo` 自算前/后复权 |

**补充 / 备选源**（均通过 `IMarketProvider` 接入，按需启用）：

| 库 / 源 | 用途定位 | 说明 |
|---------|---------|------|
| **Tushare Pro（HTTP API）** | 权威历史 / 财务 / 基本面校验 | 官方成熟，仅需 HTTP + token，无需 Go SDK；可作 gotdx 数据的交叉校验/补充 |
| Python sidecar（AKShare / BaoStock） | 海量历史 / 特色指标 | Python 生态最成熟（高 star），起一个内网小服务供 Go 调用，规避低 star Go wrapper 风险 |
| efinance-go / akshare-go 等 | 仅作参考 | star 低、维护风险高，**仅作协议/字段参考，不建议直接 import 为核心依赖** |

**推荐落地组合（V1）**：
- **实时行情 / K线 / 分时 / F10 / 财务 / 资金流 → `gotdx`（主力）**。
- **权威历史 / 基本面（可选校验）→ Tushare Pro HTTP**（用户在 M7 填自己的 token）。
- 通过配置 `MARKET_PROVIDER=gotdx|tushare` 选择实现；支持 M7 用户级覆盖。

> ⚠️ 合规提示：通达信等公开/逆向源仅适合个人非商业使用，存在频率限制与稳定性风险。务必做好限频、重试、缓存兜底；未来商用应评估合规并接入券商 QMT / 持牌数据厂商。

#### gotdx 连接池封装（解决 TDX 单连接非并发安全）

> TDX 单个 `gotdx.Client` 的 TCP 连接**不可被多 goroutine 并发复用**。后端在 Gin 高并发下必须用连接池：每次请求借一个连接，用完归还；连接断开时重建。

```go
// internal/integration/market/gotdx_pool.go
package market

import (
    "context"
    "sync"

    "github.com/bensema/gotdx"
)

type gotdxPool struct {
    mu    sync.Mutex
    idle  []*gotdx.Client
    max   int
    n     int
    newFn func() (*gotdx.Client, error)
}

func newGotdxPool(max int) *gotdxPool {
    return &gotdxPool{
        max: max,
        newFn: func() (*gotdx.Client, error) {
            hosts := gotdx.MainHostAddresses()
            cli := gotdx.New(
                gotdx.WithTCPAddress(hosts[0]),
                gotdx.WithTCPAddressPool(hosts[1:]...),
                gotdx.WithAutoSelectFastest(true), // 自动选最快节点
                gotdx.WithTimeoutSec(6),
            )
            return cli, cli.Connect()
        },
    }
}

func (p *gotdxPool) Get() (*gotdx.Client, error) {
    p.mu.Lock()
    if len(p.idle) > 0 {
        cli := p.idle[len(p.idle)-1]
        p.idle = p.idle[:len(p.idle)-1]
        p.mu.Unlock()
        return cli, nil
    }
    p.n++
    p.mu.Unlock()
    return p.newFn() // 新建连接（含 host 测速）
}

func (p *gotdxPool) Put(cli *gotdx.Client, broken bool) {
    if broken { // 连接异常则丢弃并重置计数，下次重建
        _ = cli.Disconnect()
        p.mu.Lock(); p.n--; p.mu.Unlock()
        return
    }
    p.mu.Lock(); p.idle = append(p.idle, cli); p.mu.Unlock()
}
```

#### Provider 脚手架示例（gotdx）

```go
// internal/integration/market/gotdx_provider.go
package market

import (
    "context"

    "github.com/bensema/gotdx"
    "warden/internal/model"
)

type gotdxProvider struct{ pool *gotdxPool }

func NewGotdxProvider(maxConn int) IMarketProvider {
    return &gotdxProvider{pool: newGotdxPool(maxConn)}
}

func (p *gotdxProvider) Quotes(ctx context.Context, codes []string) ([]model.StockQuote, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err() // 传播超时/取消
    default:
    }
    cli, err := p.pool.Get()
    if err != nil {
        return nil, err
    }
    rows, err := cli.StockQuotesDetail(marketsOf(codes), symbolsOf(codes))
    p.pool.Put(cli, err != nil) // 出错视为连接可能损坏，归还时丢弃
    if err != nil {
        return nil, err
    }
    return mapQuotes(rows), nil
}

func (p *gotdxProvider) Kline(ctx context.Context, code, period, adjust string) ([]model.Kline, error) {
    cli, err := p.pool.Get()
    if err != nil {
        return nil, err
    }
    bars, err := cli.StockKLine(marketOf(code), code, periodOf(period), 0, 800)
    if err != nil {
        p.pool.Put(cli, true)
        return nil, err
    }
    var xdxr []gotdx.XDXRInfo
    if adjust != "" { // 需要复权时拉除权除息数据
        xdxr, err = cli.GetXDXRInfo(marketOf(code), code)
    }
    p.pool.Put(cli, err != nil)
    if err != nil {
        return nil, err
    }
    return applyAdjust(mapKlines(bars), xdxr, adjust), nil // 自算前/后复权
}

// Indices / Search 同理封装；mapXxx / applyAdjust 负责字段映射与复权计算
```

#### 复权计算说明（M2 回测必备）

gotdx 返回的 K 线为**不复权**价。回测需用 `GetXDXRInfo`（每次分红送股/配股的除权除息记录）按口径换算：

- **前复权（qfq，默认）**：以最新价为基准，把历史价向前调整，保证最新价不变，**回测/画图常用**。
- **后复权（hfq）**：以上市首日为基准向后调整，**计算长期累计收益**用。
- 实现：按除权除息事件计算每段复权因子，逐根 K 线乘以因子；`applyAdjust(klines, xdxr, "qfq"|"hfq"|"")`。该函数为纯函数，须有单测覆盖（见 §6.2）。

> Service 层只依赖 `IMarketProvider` 接口，测试时用 mockgen 生成 mock，禁止单测中真实外呼（见 §6.2）。切换数据源仅需在依赖注入处替换 `NewGotdxProvider()` 为其他实现（如 Tushare）。

### M2 策略模块设计

- 指标条件以 JSONB（`{logic, rules[]}`）存储，回测/选股引擎解析执行。
- skill.md 每次保存 `version+1`，保留历史；AI 分析按最新版本加载。
- **回测异步化**：`POST backtest` 创建 `backtest_results(status=pending)` → 投递 MQ/启 goroutine → 引擎用历史行情逐日撮合 → 写回 `metrics/curve/status`。Service 层用 `select{case <-ctx.Done()}` 支持中断。

### M3 持仓模块设计（盈亏计算核心）

**口径：移动加权平均成本（默认）**，全程在事务中重算，保证一致性。

- 买入：`new_qty = qty + buyQty`；`new_total_cost = total_cost + price*buyQty + fee`；`avg_cost = new_total_cost / new_qty`。
- 卖出：`realized = (price - avg_cost)*sellQty - fee - tax`；`qty -= sellQty`；`total_cost -= avg_cost*sellQty`；累加 `realized_pnl`；`qty==0` 则置已清仓。
- 浮动盈亏（展示期计算）：`(currentPrice - avg_cost) * qty`，currentPrice 来自 M1。

```go
func (s *positionService) AddTrade(ctx context.Context, uid uint, posID uint, req *dto.AddTradeReq) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        pos, err := s.repo.FindByIDForUpdate(ctx, tx, uid, posID) // SELECT ... FOR UPDATE
        if err != nil { return err }
        if err := recalcPosition(pos, req); err != nil { return err } // 上述口径
        if err := s.repo.SaveTrade(ctx, tx, buildTrade(uid, posID, req)); err != nil { return err }
        return s.repo.UpdatePosition(ctx, tx, pos)
    })
}
```

### M4 风控模块设计

- 组装上下文：当前持仓 + 行情 + 市场环境 → 拼装 prompt → 调 `integration/llm` 流式输出 → 落 `position_analysis_reports`。
- 风险规则引擎：集中度/回撤/止损阈值等规则计算 `risk_items` 与 `risk_level`，与 AI 文本互补。
- 盘前预案：对每只持仓生成关注价位/止盈止损/加减仓建议，存 `premarket_plans`，可触发推送。
- 异动告警：行情轮询/任务触发时比对阈值，命中写 `alert_records` 并经飞书推送。

### M5 定时任务模块设计

- **调度器**：`scheduler` 基于 robfig/cron，启动时从 `scheduled_tasks` 加载启用任务注册；CRUD 时动态增删 cron entry。
- **交易日历**：`trading_day_only=true` 的任务在执行前判断交易日，非交易日跳过。
- **执行器**：按 `task_type` 分派（fetch_data/analysis/report/push），每次执行写 `task_execution_logs`（running→success/failed），失败可重试，支持 `POST /run` 手动触发。
- **推送**：`integration/feishu` 提供机器人 Webhook 与云文档写入，凭证从 M7 用户配置解密读取。

### M6 AI 模块设计

- 会话发起绑定 `strategy_id + stock_code`；系统消息加载该策略最新 skill.md 作为分析框架。
- 上下文注入：个股行情（M1）+ 必要持仓（M3）。
- **流式**：`integration/llm` 以 stream 方式产出 → Handler 透传 SSE。
- 报告：对话结束「生成报告」→ 结构化落 `stock_analysis_reports`，以 `(user_id, stock_code)` 建索引，`/ai/reports/grouped` 按代码聚合返回。
- 凭证：大模型 Key 由后端从 M7 配置解密使用，**绝不下发前端**。

### M7 配置模块设计

- 分组键值表，敏感字段（`is_secret=true`）经 `pkg/crypto` AES-GCM 加密存 `value_cipher`。
- 读取：返回时敏感值脱敏（`****后四位`）；内部模块（AI/飞书/数据源）通过 Service 解密获取明文使用。
- 主密钥来自环境变量 `CONFIG_ENC_KEY`，不入库不入仓。

---

## 6. TDD 驱动测试文档

### 6.1 流程与组织

- 流程：**红灯（先写失败测试）→ 绿灯（最小实现）→ 重构**。
- 单测：与源文件同目录 `xxx_test.go`；集成测试：`test/`；Mock：`internal/mock/`（mockgen 生成）。
- 工具：`testify/assert`、`go.uber.org/mock`、表驱动测试。

```bash
go install go.uber.org/mock/mockgen@latest
mockgen -source=internal/repository/position.go -destination=internal/mock/position_repository_mock.go -package=mock
```

### 6.2 测试矩阵（按模块）

| 模块 | 重点单测 | 关键用例 |
|------|---------|---------|
| M3 持仓 | `recalcPosition` 盈亏算法 | 买入均价、部分卖出已实现盈亏、清仓状态、费用税费、数量为负拒绝 |
| M2 策略 | 指标条件解析、回测撮合、**复权计算 `applyAdjust`** | 条件 and/or 组合、回测中断、参数非法、前复权/后复权/不复权口径正确 |
| M1 行情 | 缓存命中/降级 | 缓存命中不调外部、外部失败返回快照 stale |
| M4 风控 | 风险规则引擎 | 集中度超限、止损触发、等级判定 |
| M5 任务 | cron 解析、交易日跳过、执行日志 | 非交易日跳过、失败记录、手动触发 |
| M6 AI | prompt 组装、报告归档 | skill 加载、按代码归档检索 |
| M7 配置 | 加解密、脱敏 | 加密往返一致、脱敏只露后四位 |
| 鉴权 | 中间件 user_id 注入 | 单用户模式默认用户、越权拦截 |

### 6.3 Service 层表驱动测试模板（持仓盈亏）

```go
func TestPositionService_AddTrade(t *testing.T) {
    tests := []struct {
        name     string
        pos      *model.Position
        req      *dto.AddTradeReq
        wantQty  string
        wantAvg  string
        wantErr  bool
    }{
        {
            name:    "首次买入计算均价",
            pos:     &model.Position{Quantity: dec("0"), TotalCost: dec("0")},
            req:     &dto.AddTradeReq{Side: 1, Price: dec("10"), Quantity: dec("1000"), Fee: dec("5")},
            wantQty: "1000", wantAvg: "10.005", wantErr: false,
        },
        {
            name:    "超量卖出报错",
            pos:     &model.Position{Quantity: dec("100"), AvgCost: dec("10")},
            req:     &dto.AddTradeReq{Side: 2, Price: dec("11"), Quantity: dec("200")},
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := recalcPosition(tt.pos, tt.req)
            if (err != nil) != tt.wantErr {
                t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
            }
            if tt.wantErr { return }
            assert.Equal(t, tt.wantQty, tt.pos.Quantity.String())
            assert.Equal(t, tt.wantAvg, tt.pos.AvgCost.String())
        })
    }
}
```

### 6.4 Service 用 Mock Repository 测试

```go
func TestPositionService_AddTrade_RepoError(t *testing.T) {
    ctrl := gomock.NewController(t); defer ctrl.Finish()
    repo := mock.NewMockPositionRepository(ctrl)
    repo.EXPECT().FindByIDForUpdate(gomock.Any(), gomock.Any(), uint(1), uint(12)).
        Return(nil, errcode.ErrNotFound)
    svc := service.NewPositionService(repo, nil)
    err := svc.AddTrade(context.Background(), 1, 12, &dto.AddTradeReq{Side: 1})
    assert.ErrorIs(t, err, errcode.ErrNotFound)
}
```

### 6.5 Handler 层测试

```go
func TestPositionHandler_AddTrade(t *testing.T) {
    ctrl := gomock.NewController(t); defer ctrl.Finish()
    svc := mock.NewMockPositionService(ctrl)
    svc.EXPECT().AddTrade(gomock.Any(), uint(1), uint(12), gomock.Any()).Return(nil)
    h := handler.NewPositionHandler(svc)

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Set("user_id", uint(1))
    c.Params = gin.Params{{Key: "id", Value: "12"}}
    c.Request = httptest.NewRequest("POST", "/positions/12/trades",
        strings.NewReader(`{"side":1,"price":10,"quantity":1000,"traded_at":"2026-05-30T09:35:00+08:00"}`))
    c.Request.Header.Set("Content-Type", "application/json")

    h.AddTrade(c)
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### 6.6 集成测试与覆盖率

- 集成测试用 Docker 起测试用 PostgreSQL（或 `dockertest`），跑 Repository + 真实 SQL。
- 外部依赖（行情源/大模型/飞书）在测试中以接口 Mock 替换，禁止真实外呼。
- 目标覆盖率：核心 Service（持仓盈亏、回测、风控规则、配置加解密）≥ 85%；整体 ≥ 70%。
- CI 执行：`go test ./... -race -cover`。

---

## 7. 基础设施与部署

### 7.1 docker-compose（节选）

```yaml
services:
  postgres:   # postgres:16-alpine, init.sql 挂载初始化
  redis:      # redis:7-alpine, requirepass
  rabbitmq:   # rabbitmq:3-management-alpine（可选）
  backend:    # 本服务镜像，依赖以上服务 healthcheck
```

### 7.2 关键环境变量

```env
APP_PORT=8080
SINGLE_USER_MODE=true
JWT_SECRET=change_me
CONFIG_ENC_KEY=32_bytes_key_for_aes_gcm________
PG_HOST=postgres PG_PORT=5432 PG_USER=postgres PG_PASSWORD=postgres PG_DB=warden
REDIS_HOST=redis REDIS_PORT=6379 REDIS_PASSWORD=redis123
# 外部集成（也可走 M7 用户级配置覆盖）
LLM_BASE_URL= LLM_API_KEY= LLM_MODEL=
MARKET_PROVIDER=gotdx          # gotdx(主力,直连通达信)|tushare
MARKET_GOTDX_MAX_CONN=8        # gotdx 连接池大小
TUSHARE_TOKEN=                 # 仅当 MARKET_PROVIDER=tushare 或用作补充校验时
```

### 7.3 中间件装配顺序

`Recovery → Logger → CORS → RateLimit → Timeout(context) → Auth → 业务路由`。所有 DB 操作 `WithContext(ctx)`，事务回调返回 error 即回滚，保证超时/取消传播。

---

## 8. 并行开发约定（后端视角）

1. 先用 swaggo 产出/维护 OpenAPI，作为前端类型与 Mock 的来源。
2. 接口契约（响应体、分页、错误码、SSE 协议）冻结后再并行实现，变更走 PR + 版本说明。
3. 按 `internal/handler|service|repository` 分模块并行；外部集成层先以接口 + Mock 落地，后接真实数据源。
4. TDD：先补测试再实现；提交遵循 `feat(position): ...` 等约定式提交。

---

*本后端文档与《前端技术开发文档》同级，二者共同构成开发实现依据，最终需求以 [`docs/PRD.md`](./PRD.md) 为准。*
