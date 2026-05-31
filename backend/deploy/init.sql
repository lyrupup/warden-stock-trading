-- 守望者股票交易系统 · 初始化建表脚本
-- 依据 BACKEND.md §3.3。由 docker-compose 挂载到 postgres 容器初始化执行。

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

CREATE TABLE IF NOT EXISTS strategy_indicators (
    id BIGSERIAL PRIMARY KEY,
    strategy_id BIGINT NOT NULL,
    conditions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_strategy_indicators_strategy_id ON strategy_indicators(strategy_id);

CREATE TABLE IF NOT EXISTS strategy_skills (
    id BIGSERIAL PRIMARY KEY,
    strategy_id BIGINT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_strategy_skills_strategy_id ON strategy_skills(strategy_id);

CREATE TABLE IF NOT EXISTS backtest_results (
    id BIGSERIAL PRIMARY KEY,
    strategy_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    params JSONB NOT NULL DEFAULT '{}',
    status SMALLINT NOT NULL DEFAULT 0,
    metrics JSONB,
    curve JSONB,
    error_msg VARCHAR(512) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_backtest_results_strategy_id ON backtest_results(strategy_id);

-- M2 量化粗筛任务/结果
CREATE TABLE IF NOT EXISTS strategy_screen_results (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    strategy_id BIGINT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,      -- 0 pending 1 running 2 done 3 failed
    params JSONB NOT NULL DEFAULT '{}',
    universe_count INT NOT NULL DEFAULT 0,
    matched_count INT NOT NULL DEFAULT 0,
    candidates JSONB,
    trade_date DATE,
    error_msg VARCHAR(512) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_screen_results_strategy_id ON strategy_screen_results(strategy_id);
CREATE INDEX IF NOT EXISTS idx_screen_results_user_id ON strategy_screen_results(user_id);

-- M2 个股技术指标快照（公共, 可选；由 M5 盘后定时任务预计算，全市场粗筛加速）
CREATE TABLE IF NOT EXISTS stock_indicator_snapshots (
    id BIGSERIAL PRIMARY KEY,
    stock_code VARCHAR(16) NOT NULL,
    stock_name VARCHAR(64) NOT NULL DEFAULT '',
    trade_date DATE NOT NULL,
    close NUMERIC(20,4), prev_close NUMERIC(20,4),
    ma5 NUMERIC(20,4), ma10 NUMERIC(20,4), ma20 NUMERIC(20,4),
    ma30 NUMERIC(20,4), ma60 NUMERIC(20,4),
    bias5 NUMERIC(10,4), bias10 NUMERIC(10,4), bias20 NUMERIC(10,4),
    amplitude NUMERIC(10,4),
    amplitude_streak SMALLINT NOT NULL DEFAULT 0,
    turnover_rate NUMERIC(10,4), volume NUMERIC(20,0), change_percent NUMERIC(10,4),
    factors JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (stock_code, trade_date)
);
CREATE INDEX IF NOT EXISTS idx_indicator_snap_date ON stock_indicator_snapshots(trade_date);

-- M3 持仓
CREATE TABLE IF NOT EXISTS positions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    stock_code VARCHAR(16) NOT NULL,
    stock_name VARCHAR(64) NOT NULL DEFAULT '',
    quantity NUMERIC(20,4) NOT NULL DEFAULT 0,
    avg_cost NUMERIC(20,4) NOT NULL DEFAULT 0,
    total_cost NUMERIC(20,4) NOT NULL DEFAULT 0,
    realized_pnl NUMERIC(20,4) NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_positions_user_id ON positions(user_id);

CREATE TABLE IF NOT EXISTS trades (
    id BIGSERIAL PRIMARY KEY,
    position_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    side SMALLINT NOT NULL,
    price NUMERIC(20,4) NOT NULL,
    quantity NUMERIC(20,4) NOT NULL,
    fee NUMERIC(20,4) NOT NULL DEFAULT 0,
    tax NUMERIC(20,4) NOT NULL DEFAULT 0,
    realized_pnl NUMERIC(20,4) NOT NULL DEFAULT 0,
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
    content TEXT NOT NULL DEFAULT '',
    risk_items JSONB,
    risk_level SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_par_user_date ON position_analysis_reports(user_id, report_date);

CREATE TABLE IF NOT EXISTS premarket_plans (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    trade_date DATE NOT NULL,
    plans JSONB NOT NULL DEFAULT '[]',
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
    task_type VARCHAR(32) NOT NULL,
    cron_expr VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
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
    status SMALLINT NOT NULL DEFAULT 0,
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
    role VARCHAR(16) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_conv_msg_conv_id ON conversation_messages(conversation_id);

CREATE TABLE IF NOT EXISTS stock_analysis_reports (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    stock_code VARCHAR(16) NOT NULL,
    strategy_id BIGINT NOT NULL,
    conversation_id BIGINT,
    title VARCHAR(128) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    tags VARCHAR(256) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sar_user_code ON stock_analysis_reports(user_id, stock_code);

-- M7 用户配置（分组键值，敏感加密）
CREATE TABLE IF NOT EXISTS user_configs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    config_group VARCHAR(32) NOT NULL,
    config_key VARCHAR(64) NOT NULL,
    config_value TEXT NOT NULL DEFAULT '',
    value_cipher BYTEA,
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, config_group, config_key)
);
CREATE INDEX IF NOT EXISTS idx_user_configs_user_id ON user_configs(user_id);
