import { http } from "@/core/http-client";
import { toNumber } from "@/lib/decimal";
import type {
  TAddWatchParams,
  TIndexQuote,
  TKline,
  TKlinePeriod,
  TStockBrief,
  TStockQuote,
  TWatchItem,
} from "./types";

/**
 * 数值字段归一化：后端用 shopspring/decimal 序列化，数值会以 JSON 字符串返回
 * （如 "10.5"）。在 API 边界统一转成 number，保证下游 `number` 类型在运行时为真，
 * 避免组件 / 图表 / 格式化函数收到字符串而出错。
 */
function normalizeIndexQuote(q: TIndexQuote): TIndexQuote {
  return {
    ...q,
    price: toNumber(q.price),
    change_amount: toNumber(q.change_amount),
    change_percent: toNumber(q.change_percent),
    volume: toNumber(q.volume),
    amount: toNumber(q.amount),
  };
}

function normalizeStockQuote(q: TStockQuote): TStockQuote {
  return {
    ...q,
    price: toNumber(q.price),
    open: toNumber(q.open),
    high: toNumber(q.high),
    low: toNumber(q.low),
    prev_close: toNumber(q.prev_close),
    change_percent: toNumber(q.change_percent),
    volume: toNumber(q.volume),
    amount: toNumber(q.amount),
    turnover_rate: toNumber(q.turnover_rate),
  };
}

function normalizeKline(k: TKline): TKline {
  return {
    ...k,
    open: toNumber(k.open),
    high: toNumber(k.high),
    low: toNumber(k.low),
    close: toNumber(k.close),
    volume: toNumber(k.volume),
    amount: toNumber(k.amount),
  };
}

/**
 * M1 行情接口（对齐 openapi /market/*）。
 * 工厂函数注入 http，便于测试 Mock（见 FRONTEND.md §4.2）。
 */
export function createMarketApi(client = http) {
  return {
    /** 当日大盘指数列表 */
    indices: () =>
      client.get<TIndexQuote[]>("market/indices").then((list) => list.map(normalizeIndexQuote)),
    /** 自选股列表 */
    watchlist: () => client.get<TWatchItem[]>("market/watchlist"),
    /** 自选股实时行情 */
    watchlistQuotes: () =>
      client
        .get<TStockQuote[]>("market/watchlist/quotes")
        .then((list) => list.map(normalizeStockQuote)),
    /** 添加自选股 */
    addWatch: (data: TAddWatchParams) => client.post<void>("market/watchlist", { json: data }),
    /** 删除自选股 */
    removeWatch: (id: number) => client.delete<void>(`market/watchlist/${id}`),
    /** 个股行情详情 */
    stock: (code: string) =>
      client.get<TStockQuote>(`market/stocks/${code}`).then(normalizeStockQuote),
    /** K 线数据 */
    kline: (code: string, period: TKlinePeriod = "day", adjust: "" | "qfq" | "hfq" = "qfq") =>
      client
        .get<TKline[]>(`market/stocks/${code}/kline`, {
          searchParams: { period, adjust },
        })
        .then((list) => list.map(normalizeKline)),
    /** 按代码/名称搜索股票 */
    search: (kw: string) =>
      client.get<TStockBrief[]>("market/search", { searchParams: { kw } }),
  };
}

export const marketApi = createMarketApi();

export type TMarketApi = ReturnType<typeof createMarketApi>;
