import { http } from "@/core/http-client";
import type {
  TAddWatchParams,
  TIndexQuote,
  TKline,
  TKlinePeriod,
  TStockQuote,
  TWatchItem,
} from "./types";

/**
 * M1 行情接口（对齐 openapi /market/*）。
 * 工厂函数注入 http，便于测试 Mock（见 FRONTEND.md §4.2）。
 */
export function createMarketApi(client = http) {
  return {
    /** 当日大盘指数列表 */
    indices: () => client.get<TIndexQuote[]>("market/indices"),
    /** 自选股列表 */
    watchlist: () => client.get<TWatchItem[]>("market/watchlist"),
    /** 自选股实时行情 */
    watchlistQuotes: () => client.get<TStockQuote[]>("market/watchlist/quotes"),
    /** 添加自选股 */
    addWatch: (data: TAddWatchParams) => client.post<void>("market/watchlist", { json: data }),
    /** 删除自选股 */
    removeWatch: (id: number) => client.delete<void>(`market/watchlist/${id}`),
    /** 个股行情详情 */
    stock: (code: string) => client.get<TStockQuote>(`market/stocks/${code}`),
    /** K 线数据 */
    kline: (code: string, period: TKlinePeriod = "day", adjust: "" | "qfq" | "hfq" = "qfq") =>
      client.get<TKline[]>(`market/stocks/${code}/kline`, {
        searchParams: { period, adjust },
      }),
  };
}

export const marketApi = createMarketApi();

export type TMarketApi = ReturnType<typeof createMarketApi>;
