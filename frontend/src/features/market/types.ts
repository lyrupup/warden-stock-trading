/** 指数行情（对应 openapi IndexQuote） */
export type TIndexQuote = {
  index_code: string;
  index_name: string;
  price: number;
  change_amount: number;
  change_percent: number;
  volume: number;
  amount: number;
  trade_date: string;
};

/** 个股行情（对应 openapi StockQuote） */
export type TStockQuote = {
  stock_code: string;
  stock_name: string;
  price: number;
  open: number;
  high: number;
  low: number;
  prev_close: number;
  change_percent: number;
  volume: number;
  amount: number;
  turnover_rate: number;
  trade_date: string;
  /** 是否为降级返回的过期快照 */
  stale?: boolean;
};

/** 自选股（对应 openapi WatchItem） */
export type TWatchItem = {
  id: number;
  stock_code: string;
  stock_name: string;
  group_name: string;
  remark?: string;
};

/** 添加自选股参数 */
export type TAddWatchParams = {
  stock_code: string;
  group_name?: string;
  remark?: string;
};

/** K 线周期（对应 openapi kline period 枚举） */
export type TKlinePeriod = "day" | "week" | "month" | "1m" | "5m" | "15m" | "30m" | "60m";

/** K 线数据（对应 openapi Kline） */
export type TKline = {
  date: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  amount: number;
};
