# 守望者股票交易系统 · 前端技术开发文档

> Warden Stock Trading System · Frontend Technical Design
>
> 文档版本：v1.0　|　创建日期：2026-05-30　|　依据：[`docs/PRD.md`](./PRD.md)
>
> 适用范围：本文档为前端团队的开发基线，与《后端技术开发文档》并行开发，二者通过《API 接口文档》（见后端文档）对齐契约。

---

## 1. 技术栈与选型

### 1.1 主要技术栈

| 分类 | 选型 | 说明 |
|------|------|------|
| 框架 | **React 18 + TypeScript** | 函数式组件 + Hooks，全量 TS |
| 构建 | **Vite 5** | 快速冷启动与 HMR |
| UI 组件 | **shadcn/ui** | 基于 Radix + Tailwind，CLI 安装到 `components/ui/` |
| 样式 | **Tailwind CSS 3** | utility-first，支持 Light/Dark 主题 |
| 路由 | **React Router v6** | 集中式路由配置于 `routes/` |
| 服务端状态 | **TanStack Query v5** | 请求缓存、轮询、重试、失效 |
| 全局状态 | **zustand** | 认证、主题、用户配置等跨组件状态 |
| 请求器 | **ky** | 基于 fetch，封装于 `core/http-client/` |
| 表单 | **react-hook-form + zod** | 表单状态 + schema 校验 |
| 图表 | **lightweight-charts**（K 线/分时）+ **ECharts**（回测/统计） | 行情专业图与通用统计图分工 |
| Markdown | **react-markdown + remark-gfm** | 策略 skill.md、AI 报告渲染 |
| 编辑器 | **@uiw/react-md-editor** 或 CodeMirror | skill.md 在线编辑 |
| 国际化 | **i18next + react-i18next** | 语言包置于 `core/i18n/locales/` |
| 测试 | **Vitest + Testing Library + MSW** | 单测 + 组件测试 + 接口 Mock |
| 代码规范 | ESLint + Prettier | 统一风格 |

### 1.2 强制编码约束（贯穿全项目）

1. **具名导出**：禁止 `export default`，统一 `export`。
2. **类型前缀**：`type` 用 `T`、`enum` 用 `E`、`interface` 用 `I`。
3. **目录命名**：组件/hook 目录用 **kebab-case**，通过 `index.ts` 统一导出。
4. **分层**：`Component → Hook/Store → Core → Backend`，禁止跨层直连。
5. **DRY**：相同/相似逻辑出现 ≥ 3 次必须抽象到 `components/common/` 或 `hooks/`。
6. **样式**：优先 Tailwind utility class，用 `cn()` 合并 className。
7. **金额/股价**：统一通过工具函数格式化（精度、涨跌色），禁止散落处理。

---

## 2. 项目目录结构

```
src/
├── components/
│   ├── ui/                      # shadcn/ui 原始组件（CLI 生成，勿改）
│   └── common/                  # 业务通用组件（基于 ui/ 封装）
│       ├── data-table/          # 通用表格（排序/分页/列配置）
│       ├── quote-cell/          # 行情涨跌色单元格
│       ├── confirm-dialog/      # 二次确认弹窗
│       ├── markdown-viewer/     # Markdown 渲染（AI 报告/skill）
│       ├── markdown-editor/     # skill.md 编辑器
│       ├── page-header/         # 页面标题/操作区
│       └── empty-state/         # 空态/错误降级
├── features/                    # 业务功能模块（按业务域划分）
│   ├── auth/                    # 登录/鉴权（多用户预留）
│   ├── market/                  # M1 行情数据
│   ├── strategy/                # M2 选股策略
│   ├── position/                # M3 持仓记录
│   ├── risk/                    # M4 持仓分析与风控
│   ├── task/                    # M5 定时任务
│   ├── ai-analysis/             # M6 AI 对话分析
│   └── settings/                # M7 个人中心配置
├── core/
│   ├── http-client/             # ky 封装 + 拦截器
│   ├── api/                     # 各业务 service 工厂（基于 http-client）
│   ├── auth/                    # token 管理
│   ├── i18n/                    # i18next + locales/
│   ├── theme/                   # 主题（Light/Dark）provider
│   └── ws/                      # 行情推送（SSE/WebSocket）封装
├── hooks/                       # 全局共享 hooks（kebab-case）
│   ├── use-paged-query/
│   └── use-debounce/
├── stores/                      # zustand 全局 store
│   ├── auth-store.ts
│   ├── theme-store.ts
│   └── user-config-store.ts
├── types/                       # 全局类型（与后端 DTO 对齐）
├── lib/                         # 工具函数（cn、格式化、日期）
├── routes/                      # 路由配置 + 守卫
├── styles/                      # Tailwind 扩展、全局样式
├── App.tsx
└── main.tsx
```

每个 `features/<module>/` 内部统一结构：

```
features/strategy/
├── components/        # 模块专属组件
├── hooks/             # 模块专属 hooks（封装 TanStack Query）
├── api.ts             # 模块接口（调用 core/api 或直接基于 http-client）
├── types.ts           # 模块类型（TStrategy 等）
└── index.ts           # 统一具名导出
```

---

## 3. 核心功能包（依赖清单）

| 包 | 用途 | 所在层 |
|----|------|-------|
| `react` / `react-dom` | 框架 | 全局 |
| `vite` / `@vitejs/plugin-react` | 构建 | 工程 |
| `tailwindcss` / `postcss` / `autoprefixer` | 样式 | 全局 |
| `class-variance-authority` / `clsx` / `tailwind-merge` | shadcn 样式合并（`cn`） | `lib/` |
| `@radix-ui/*`（随 shadcn 安装） | 无障碍基础组件 | `components/ui/` |
| `react-router-dom` | 路由 | `routes/` |
| `@tanstack/react-query` | 服务端状态 | `features/*/hooks` |
| `zustand` | 全局状态 | `stores/` |
| `ky` | HTTP 请求 | `core/http-client/` |
| `react-hook-form` / `zod` / `@hookform/resolvers` | 表单与校验 | `features/*` |
| `lightweight-charts` | K 线/分时图 | `features/market` |
| `echarts` / `echarts-for-react` | 回测/统计图 | `features/strategy`、`features/position` |
| `react-markdown` / `remark-gfm` | Markdown 渲染 | `components/common/markdown-viewer` |
| `@uiw/react-md-editor` | skill.md 编辑 | `components/common/markdown-editor` |
| `i18next` / `react-i18next` | 国际化 | `core/i18n/` |
| `dayjs` | 日期/交易日处理 | `lib/` |
| `vitest` / `@testing-library/react` / `msw` | 测试 | 工程 |

---

## 4. 核心基础设施实现

### 4.1 HTTP 请求器（core/http-client）

统一 ky 实例，处理 token 注入、统一响应解包（`{code,message,data}`）、401 登出、业务错误抛 `AppError`。

```typescript
// core/http-client/http-client.ts
import ky from "ky";
import { useAuthStore } from "@/stores/auth-store";

export class AppError extends Error {
  constructor(public code: number, message: string) {
    super(message);
    this.name = "AppError";
  }
}

export const httpClient = ky.create({
  prefixUrl: import.meta.env.VITE_API_BASE_URL,
  timeout: 30_000,
  hooks: {
    beforeRequest: [
      (request) => {
        const token = useAuthStore.getState().token;
        if (token) request.headers.set("Authorization", `Bearer ${token}`);
      },
    ],
    afterResponse: [
      async (_req, _opts, response) => {
        if (response.status === 401) {
          useAuthStore.getState().logout();
          return;
        }
        if (response.ok) {
          const body = await response.clone().json<{ code: number; message: string }>();
          if (body.code !== 0) throw new AppError(body.code, body.message);
        }
      },
    ],
  },
});
```

> AI 流式分析接口（M6）走 SSE/fetch-stream，单独在 `core/http-client/stream.ts` 封装，不经过上面的 JSON 解包。

### 4.2 服务工厂（core/api）

每个业务模块提供工厂函数，注入 ky 实例，返回纯接口对象（依赖注入，便于测试 Mock）。

```typescript
// features/strategy/api.ts
import { httpClient } from "@/core/http-client";
import type { TStrategy, TCreateStrategyParams } from "./types";

export function createStrategyApi(api = httpClient) {
  return {
    list: () => api.get("strategies").json<TStrategy[]>(),
    detail: (id: string) => api.get(`strategies/${id}`).json<TStrategy>(),
    create: (data: TCreateStrategyParams) => api.post("strategies", { json: data }).json<TStrategy>(),
    update: (id: string, data: Partial<TCreateStrategyParams>) =>
      api.put(`strategies/${id}`, { json: data }).json<TStrategy>(),
    remove: (id: string) => api.delete(`strategies/${id}`).json<void>(),
    // M2 量化粗筛
    indicatorCatalog: () => api.get("strategies/indicators/catalog").json<TIndicatorCatalogItem[]>(),
    templates: () => api.get("strategies/templates").json<TStrategyTemplate[]>(),
    runScreen: (id: string, params: TScreenParams) =>
      api.post(`strategies/${id}/screen`, { json: params }).json<{ taskId: string }>(),
    screenResult: (id: string, taskId: string) =>
      api.get(`strategies/${id}/screen/${taskId}`).json<TScreenResult>(),
    previewScreen: (params: TScreenParams & { indicators: TIndicatorGroup }) =>
      api.post("strategies/screen/preview", { json: params }).json<TScreenResult>(),
    runBacktest: (id: string, params: object) =>
      api.post(`strategies/${id}/backtest`, { json: params }).json<{ taskId: string }>(),
  };
}
```

### 4.3 通用分页 Hook（hooks/use-paged-query）

抽象列表页通用「分页 + 加载 + 错误」逻辑，供持仓流水、报告列表、任务日志等复用。

```typescript
// hooks/use-paged-query/use-paged-query.ts
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";

export type TPageResult<T> = { list: T[]; total: number; page: number; size: number };

export function usePagedQuery<T>(
  key: unknown[],
  fetcher: (page: number, size: number) => Promise<TPageResult<T>>,
) {
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(20);
  const query = useQuery({ queryKey: [...key, page, size], queryFn: () => fetcher(page, size) });
  return { ...query, page, size, setPage, setSize };
}
```

### 4.4 全局状态（zustand）

```typescript
// stores/auth-store.ts —— 多用户预留
type TAuthState = {
  token: string | null;
  userId: string | null;
  setAuth: (token: string, userId: string) => void;
  logout: () => void;
};
export const useAuthStore = create<TAuthState>((set) => ({
  token: localStorage.getItem("token"),
  userId: null,
  setAuth: (token, userId) => { localStorage.setItem("token", token); set({ token, userId }); },
  logout: () => { localStorage.removeItem("token"); set({ token: null, userId: null }); },
}));
```

- `theme-store`：管理 `light | dark`，写入 `localStorage` 并切换 `<html class="dark">`。
- `user-config-store`：缓存 M7 个人配置（脱敏后的非敏感部分），供全局读取。

### 4.5 主题（core/theme）

- Tailwind `darkMode: "class"`，主题色通过 CSS 变量定义在 `styles/`。
- 涨跌色遵循 A 股习惯：**涨=红、跌=绿**，在 Light/Dark 下各定义一套变量；统一通过 `getQuoteColor(changePercent)` 工具返回 class。

---

## 5. 路由与页面规划

### 5.1 路由表

| 路径 | 页面 | 模块 | 守卫 |
|------|------|------|------|
| `/login` | 登录页 | auth | 公开 |
| `/` → `/dashboard` | 工作台/首页 | 聚合 | 需登录 |
| `/market` | 行情中心（大盘指数 + 自选股） | M1 | 需登录 |
| `/market/quote` | 个股行情（搜索个股看当日行情 + K 线，左侧「侦查」独立入口） | M1 | 需登录 |
| `/market/:code` | 个股详情（与个股行情页复用同一面板） | M1 | 需登录 |
| `/strategies` | 策略列表 | M2 | 需登录 |
| `/strategies/:id` | 策略详情/编辑（含 skill.md、回测） | M2 | 需登录 |
| `/positions` | 持仓列表 | M3 | 需登录 |
| `/positions/:id` | 持仓详情（交易流水） | M3 | 需登录 |
| `/risk` | 持仓分析与风控 | M4 | 需登录 |
| `/risk/premarket` | 明日盘前预案 | M4 | 需登录 |
| `/tasks` | 定时任务管理 | M5 | 需登录 |
| `/ai` | AI 对话分析 | M6 | 需登录 |
| `/ai/reports` | 分析报告库（按股票代码归档） | M6 | 需登录 |
| `/settings` | 个人中心配置 | M7 | 需登录 |

> **多用户预留**：登录守卫读取 `auth-store.token`；V1 可在 `core/auth` 中提供「默认用户自动登录」开关（`VITE_SINGLE_USER_MODE=true` 时自动注入默认 token），不改动路由结构。

### 5.2 全局布局

- `AppLayout`：左侧导航（按四大技能彩蛋分区：侦查 / 洞察 / 警戒 / 传信 + 设置）、顶部栏（搜索股票、主题切换、用户菜单）、内容区。
- 内容区统一用 `PageHeader` + 主体，错误/空态用 `EmptyState`。

---

## 6. 各页面核心功能与实现方案

### M1 行情中心（features/market）

**页面**：行情中心 `/market`、**个股行情 `/market/quote`**（左侧「侦查」分区独立入口）、个股详情 `/market/:code`

**核心功能**
- 行情中心 `/market`：
  - 大盘指数卡片区：进入自动加载当日指数（上证/深证/创业板/科创50/沪深300），涨跌色展示。
  - 市场概览：涨跌家数、涨停跌停、北向资金（按数据源裁剪）。
  - 自选股表格：`data-table` 展示自选股实时行情，支持分组、搜索、增删、跳转详情。
- 个股行情 `/market/quote`（`pages/stock-quote-page.tsx`）：输入股票代码 / 名称搜索（`GET /market/search`）→ 选中结果后展示该股**当日行情 + 历史 K 线**；唯一匹配（如精确代码）自动选中。
- 个股详情 / 个股行情面板：`lightweight-charts` 渲染 K 线（日/周/月切换）+ 分时，基础财务概要。
  - 当日行情卡片与 K 线抽为 **`components/stock-quote-panel.tsx`**，由 `/market/:code` 详情页与 `/market/quote` 个股行情页 **共用同一组件**（按 `code` 自取数，react-query 缓存）。

**实现要点**
- 行情刷新：交易时段内用 TanStack Query 的 `refetchInterval`（频率读 M7 配置，默认 5s）；非交易时段停止轮询。可选升级为 `core/ws` 的 SSE 推送。
- 交易时段判断：`lib/trading-time.ts` 基于交易日历 + 当前时间，避免无效轮询。
- 降级：数据源异常时表格保留上次快照 + 顶部提示条，不阻塞页面。

```typescript
// features/market/hooks/use-watchlist-quotes.ts
export function useWatchlistQuotes() {
  const refreshMs = useUserConfigStore((s) => s.marketRefreshMs ?? 5000);
  const inTrading = useIsTradingTime();
  return useQuery({
    queryKey: ["market", "watchlist-quotes"],
    queryFn: () => marketApi.watchlistQuotes(),
    refetchInterval: inTrading ? refreshMs : false,
  });
}
```

### M2 选股策略（features/strategy）

**页面**：策略列表 `/strategies`、策略详情 `/strategies/:id`

**核心功能**
- 策略 CRUD + 复制 + 标签分类 + 搜索；支持从**内置模板**（短线均线多头 / 中长线震荡）一键创建。
- 策略详情分 Tab：**基本信息/描述**（Markdown 编辑）、**量化指标定义**（因子条件构造器）、**选股粗筛**（本期核心，股票池选择 + 运行 + 候选结果表）、**skill.md**（Markdown 编辑器 + 版本记录）、**回测**（后续迭代）。
- **量化指标构造器**：可视化增删规则行，每行为「左因子 · 操作符 · 右值（因子或常量）」，可选因子来自 `/strategies/indicators/catalog`（按因子动态渲染参数：均线周期、排列周期组、乖离周期、振幅阈值/连续天数等），支持「与/或」分组嵌套，输出与后端一致的 `TIndicatorGroup` JSON。
  - **参数标签前缀**：每个参数输入框前以灰色短标签（`周期 / 周期组 / 方向 / 容差 / 阈值% / 天数 / 字段 / 值`）提示语义，长文描述走 `title` tooltip（取自 catalog `desc`）。映射表 `paramLabel(key)` 维护在 `factor-utils.ts`，避免裸数字/枚举令非开发用户无法识别。
  - **方向枚举文案**：`bull / bear` 在构造器中显示为「多头（严格递减）/ 空头（严格递增）」（`directionLabel`），与 `describeFactor` 的多/空表述对齐。后端 enum 仍是 `bull/bear`，仅前端翻译。
  - **K 线字段枚举文案**：`close/open/...` 通过 `fieldLabel` 渲染为「收盘价/开盘价/最高价/最低价/前收盘/成交量/成交额/涨跌幅」，因子名「原始字段」覆盖为「K 线字段（开高低收等）」。
- **选股粗筛（F2.6）**：
  - 股票池选择器：全市场 / 指定板块 / 自选股 / 自定义代码（`TScreenUniverse`）。
  - 「运行粗筛」→ 异步任务，轮询状态（pending/running/done/failed），完成后用 `data-table` 展示**候选列表**（代码、名称、命中因子快照值如 MA5/MA10/MA20/乖离/振幅、匹配评分），命中规则高亮。
  - 「快速预览」：自选/小池走同步 `screen/preview`，便于边调参边看效果。
  - 候选行操作：**加入自选**（复用 M1 接口）、**发起 AI 分析**（跳转 M6，携带 `strategyId + stockCode` 完成「粗筛 → AI 精筛」交接）。

**实现要点**
- 指标定义类型 `TIndicatorGroup`/`TIndicatorRule`/`TFactorExpr`，前后端共享同一 schema（zod 定义并导出类型，对齐 openapi `IndicatorGroup`），构造器渲染、提交、粗筛三处共用。
- 因子目录用 `useIndicatorCatalog`（TanStack Query，长缓存 `staleTime`），驱动构造器可选项与参数表单。
- 粗筛任务异步：`runScreen` 返回 `taskId` → `use-screen-result` 轮询 `screen/:taskId` 直至 `done/failed`；候选量大时表格分页 + 排序（评分/各因子列）。
- 候选因子快照为 decimal 字符串，经 `lib/decimal.ts` 的 `toNumber`/`coerceDecimalFields` 转换后再格式化展示，涨跌/乖离正负用 `getQuoteColor` 着色。
- **因子展示分类**（`candidate-table.tsx` 中的 `factorMeta`）：候选表按因子 key 推断显示类型，避免布尔 / 百分比 / 倍数被一锅炖成无单位小数：
  - `bool`（`ma_align_*`、`amplitude_streak`）：后端用 1/0 表示，UI 渲染为 ✓/✗，✓ 走 `text-quote-up`、✗ 走 `text-muted-foreground`；列居中。
  - `percent`（`bias*`、`amplitude`、`pct_change_*`）：后端值已乘 100，UI 追加 `%` 后缀（如 `7.72%`）。
  - `ratio`（`vol_ratio_*`）：保留两位小数、不加单位。
  - `number`（`ma*`、`close` 等）：`formatPrice(v, 2)` 渲染。
  排序按 accessor 原始数值，不会被显示文本（✓/✗ 或 `%` 后缀）扰动。
- skill.md 编辑用 `markdown-editor`，保存生成新版本（前端记录版本号，内容存后端）。
- 模块结构：`features/strategy/{components/indicator-builder, components/screen-panel, components/candidate-table, hooks/use-indicator-catalog, hooks/use-screen-result, api.ts, types.ts}`。

### M3 持仓记录（features/position）

**页面**：持仓列表 `/positions`、持仓详情 `/positions/:id`

**核心功能**
- 账户总览卡片：总市值、总成本、总浮盈、总收益率、当日盈亏。
- 持仓列表：每只股票的数量、均价、现价、浮动盈亏、收益率（行情联动刷新）。
- 持仓详情：交易流水（买/卖记录），新增交易表单（价格、数量、费用、时间、备注），自动重算均价与盈亏。

**实现要点**
- 盈亏由后端计算并返回（口径见后端文档，默认移动加权平均），前端只负责展示与格式化。
- 现价来自 M1 行情，前端将持仓股票代码并入行情订阅，市值/浮盈实时刷新。
- 新增交易用 `react-hook-form + zod`，提交成功后 `invalidateQueries(['positions'])` 刷新。

### M4 持仓分析与风控（features/risk）

**页面**：风控分析 `/risk`、盘前预案 `/risk/premarket`

**核心功能**
- 一键/查看「每日持仓 AI 分析」：结构分析、集中度、行业分布、个股点评（Markdown 渲染）。
- 风险提示列表：风险项 + 风险等级（高/中/低）色标 + 建议。
- 明日盘前预案：按个股展示关注价位、止盈止损、加减仓思路，可保存/推送飞书。
- 历史报告：按日期回看分析与预案。

**实现要点**
- AI 分析可能耗时，使用流式渲染（`core/http-client/stream`）+ 骨架屏。
- 风险等级用统一 `RiskBadge` 组件；报告用 `markdown-viewer` 渲染。
- 「推送飞书」复用 M5 推送接口。

### M5 定时任务（features/task）

**页面**：任务管理 `/tasks`

**核心功能**
- 任务 CRUD + 启用/停用开关 + 手动「立即运行」。
- 调度配置：Cron 表达式输入（带可视化预览「下次执行时间」）或周期模板（每日盘前/盘后等）。
- 任务类型选择：数据拉取/分析报告/报告生成/飞书推送，按类型动态渲染参数表单。
- 执行记录抽屉：展示历史执行时间、状态、输出/错误，失败可重试。

**实现要点**
- Cron 下次执行时间用 `cronstrue` + `cron-parser` 本地预览。
- 任务表单按类型分支（discriminated union 类型 `TTaskConfig`）。
- 执行日志用 `use-paged-query` 分页加载。

### M6 AI 对话分析（features/ai-analysis）

**页面**：AI 分析对话 `/ai`、报告库 `/ai/reports`

**核心功能**
- 发起分析：选择「本人某个策略」+ 输入/选择个股 → 开始对话。
- 流式对话区：用户提问 + AI 流式回答，AI 按所选策略 skill.md 框架输出。
- 生成结构化报告：对话可「生成报告」，按 skill 约定格式渲染。
- 报告库：按**股票代码**聚合归档，支持按代码/策略/日期检索，支持导出与推送飞书。

**实现要点**
- 对话消息流用 SSE；消息状态机 `pending → streaming → done/error`。
- 报告以股票代码为主索引：报告库页用「代码分组」视图（每个代码下挂多份报告）。
- 选用的策略、模型凭证来自 M2 与 M7（前端只传 strategyId，密钥在后端读取，不下发到前端）。

### M7 个人中心配置（features/settings）

**页面**：个人中心 `/settings`

**核心功能**
- 分组配置：**AI 模型**（供应商/API Key/Base URL/模型名/温度，支持多套）、**飞书**（机器人 Webhook、App ID/Secret、云文档配置）、**数据源**（来源/Key/刷新频率）、**偏好**（主题、语言、首页默认视图）、**个人资料**（昵称/头像，多用户预留）。
- 密钥脱敏展示（仅显示后四位），保存后即时生效。
- 主题切换即时预览并持久化。

**实现要点**
- 敏感字段提交后，回显由后端返回脱敏值；未修改则不回传明文。
- 配置保存成功后更新 `user-config-store`，行情刷新频率、主题等全局即时生效。
- 表单分 Tab + `react-hook-form`，每组独立保存。

---

## 7. 与后端的契约对齐

- **统一响应体**：`{ code: number, message: string, data: T }`，`code===0` 为成功，否则 `afterResponse` 抛 `AppError`。
- **分页响应**：`{ list, total, page, size }`，对应 `TPageResult<T>`。
- **鉴权**：`Authorization: Bearer <token>`，401 触发登出。
- **类型同步**：`types/` 与后端 DTO 字段保持一致；建议后端提供 OpenAPI，前端可用脚本生成类型（可选）。
- **流式接口**：M4/M6 的 AI 接口走 SSE，约定 `event: message` 增量、`event: done` 结束、`event: error` 错误。
- **时间字段**：统一 ISO8601 字符串，前端用 dayjs 处理。

---

## 8. 测试策略

| 层级 | 工具 | 范围 |
|------|------|------|
| 工具/纯函数 | Vitest | 金额/涨跌色/交易时段/盈亏格式化等 `lib/` |
| Hook | Vitest + Testing Library | `use-paged-query`、各模块查询 hook（MSW Mock 接口） |
| 组件 | Testing Library | 关键交互组件（指标构造器、交易表单、Cron 预览） |
| 接口 Mock | MSW | 模拟后端，前端可脱离后端独立开发联调 |

- 测试文件与源文件同目录：`xxx.test.ts(x)`。
- MSW handlers 置于 `src/mocks/`，开发环境可开启，保证**前后端并行开发**期间前端不被后端阻塞。

---

## 9. 环境变量

```env
# .env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_SINGLE_USER_MODE=true        # V1 单用户模式自动登录
VITE_SSE_PATH=/api/ai/stream      # AI 流式接口路径
```

---

## 10. 并行开发约定（前端视角）

1. 以本文件 §7 契约 + 后端《API 接口文档》为准；接口未就绪时用 MSW Mock 数据开发。
2. 类型定义以后端 DTO 为单一事实源；前端在 `types/` 维护镜像类型，变更通过 PR 同步。
3. 模块按 `features/` 解耦，可分人并行（M1/M3 偏数据展示、M2/M6 偏交互、M4/M5 偏流程）。
4. 提交遵循 `feat(market): ...` 等约定式提交。

---

*本前端文档与《后端技术开发文档》同级，二者共同构成开发实现依据，最终需求以 [`docs/PRD.md`](./PRD.md) 为准。*
