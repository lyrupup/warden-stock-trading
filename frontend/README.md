# 守望者股票交易系统 · 前端

> Warden Stock Trading System · Frontend
>
> 依据 [`docs/FRONTEND.md`](../docs/FRONTEND.md)、[`docs/PRD.md`](../docs/PRD.md)、[`docs/openapi.yaml`](../docs/openapi.yaml) 搭建的前端工程骨架。

## 技术栈

React 18 + TypeScript + Vite 5 + Tailwind CSS 3（shadcn/ui 风格）+ React Router v6 + TanStack Query v5 + zustand + ky + react-hook-form/zod + i18next + dayjs + lightweight-charts；测试用 Vitest + Testing Library + MSW。

## 环境要求

- **Node.js >= 20.17**（Vite 5 / 部分依赖要求；推荐 Node 20 LTS 或 22）。本机若默认是旧版本（如 v14），请用 nvm 切换：`nvm use 22`。

## 快速开始

```bash
cd frontend
cp .env.example .env        # 按需修改
npm install
npm run dev                 # 开发（默认开启 MSW Mock，可脱离后端运行）
```

常用脚本：

| 命令 | 说明 |
|------|------|
| `npm run dev` | 启动开发服务器（Vite） |
| `npm run build` | 类型检查（tsc -b）+ 生产构建（vite build） |
| `npm run preview` | 预览构建产物 |
| `npm run test` | 运行单元 / 组件测试（vitest run） |
| `npm run test:watch` | 监听模式测试 |

> 首次启用浏览器 Mock 前需生成 service worker：`npx msw init public/ --save`（脚手架已预置 `public/mockServiceWorker.js`）。

## 环境变量

见 [`.env.example`](./.env.example)：

- `VITE_API_BASE_URL`：后端 API 基地址。
- `VITE_SINGLE_USER_MODE`：单用户模式，`true` 时守卫自动登录默认 token。
- `VITE_DEFAULT_TOKEN`：单用户模式注入的默认 token。
- `VITE_SSE_PATH`：AI 流式接口路径。
- `VITE_ENABLE_MOCK`：`true` 时开发环境启用 MSW Mock。

## 目录结构（FRONTEND.md §2）

```
src/
├── components/{ui,common}      # ui 原子组件 + 业务通用组件（data-table/quote-cell/...）
├── features/                   # 业务模块（market 为 M1 竖切范式；其余为占位页）
├── core/{http-client,api,auth,i18n,theme}  # 核心基础设施
├── hooks/{use-paged-query,use-debounce}    # 全局共享 hooks
├── stores/                     # zustand（auth/theme/user-config）
├── types/                      # 全局类型（与后端 DTO / openapi 对齐）
├── lib/                        # cn / 格式化 / 交易时段 / dayjs
├── routes/                     # 路由配置 + 守卫 + AppLayout
├── mocks/                      # MSW handlers
├── styles/                     # Tailwind 全局样式与 CSS 变量
├── App.tsx
└── main.tsx
```

## 编码约束（强制）

1. 具名导出，禁止 `export default`。
2. 类型前缀：`type` 用 `T`、`enum` 用 `E`、`interface` 用 `I`。
3. 组件 / hook 目录用 kebab-case，经 `index.ts` 统一导出。
4. 分层：`Component → Hook/Store → Core → Backend`，禁止跨层直连。
5. 样式优先 Tailwind，用 `cn()` 合并 className。
6. 金额 / 涨跌色统一走 `lib/format`（A 股涨红跌绿，`getQuoteColor`）。

## M1 行情竖切（团队范式）

`features/market/` 是一条打通的竖切，可作为后续模块的参考范式：

- `types.ts`：`TStockQuote`/`TIndexQuote`/`TWatchItem`（对齐 openapi）。
- `api.ts`：工厂函数 `createMarketApi(http)`，基于 `core/http-client`。
- `hooks/`：TanStack Query 封装，行情轮询由交易时段（`lib/trading-time`）+ M7 刷新频率控制。
- `components/`：`IndexCard`、`WatchlistTable`（基于通用 `data-table`）、`KlineChart`。
- `pages/`：`/market` 行情中心、`/market/:code` 个股详情。

## 测试

- 工具函数：`lib/format`、`lib/trading-time`。
- 通用 hook：`use-paged-query`。
- 行情：`use-watchlist-quotes`（MSW mock）、`WatchlistTable` 组件。

```bash
npm run test
```

> ⚠️ 风险提示：本系统提供的 AI 分析、风险提示、盘前预案等均为辅助决策信息，不构成投资建议。
