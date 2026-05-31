import type { TIndexQuote, TKline, TStockQuote, TWatchItem } from "@/features/market/types";

export const mockIndices: TIndexQuote[] = [
  { index_code: "000001", index_name: "上证指数", price: 3052.34, change_amount: 12.45, change_percent: 0.41, volume: 28500000000, amount: 345600000000, trade_date: "2026-05-29" },
  { index_code: "399001", index_name: "深证成指", price: 9876.21, change_amount: -34.12, change_percent: -0.34, volume: 31200000000, amount: 412300000000, trade_date: "2026-05-29" },
  { index_code: "399006", index_name: "创业板指", price: 1912.88, change_amount: 5.67, change_percent: 0.3, volume: 12300000000, amount: 178900000000, trade_date: "2026-05-29" },
  { index_code: "000688", index_name: "科创50", price: 845.12, change_amount: -2.31, change_percent: -0.27, volume: 4500000000, amount: 67800000000, trade_date: "2026-05-29" },
  { index_code: "000300", index_name: "沪深300", price: 3567.9, change_amount: 8.9, change_percent: 0.25, volume: 9800000000, amount: 234500000000, trade_date: "2026-05-29" },
];

export let mockWatchlist: TWatchItem[] = [
  { id: 1, stock_code: "600519", stock_name: "贵州茅台", group_name: "default", remark: "" },
  { id: 2, stock_code: "000858", stock_name: "五粮液", group_name: "default", remark: "" },
  { id: 3, stock_code: "300750", stock_name: "宁德时代", group_name: "default", remark: "" },
];

export const mockQuotes: Record<string, TStockQuote> = {
  "600519": { stock_code: "600519", stock_name: "贵州茅台", price: 1689.5, open: 1672.0, high: 1695.0, low: 1668.0, prev_close: 1675.0, change_percent: 0.87, volume: 2150000, amount: 3630000000, turnover_rate: 0.17, trade_date: "2026-05-29" },
  "000858": { stock_code: "000858", stock_name: "五粮液", price: 142.3, open: 144.0, high: 144.5, low: 141.8, prev_close: 143.8, change_percent: -1.04, volume: 18500000, amount: 2640000000, turnover_rate: 0.48, trade_date: "2026-05-29" },
  "300750": { stock_code: "300750", stock_name: "宁德时代", price: 198.76, open: 196.0, high: 200.1, low: 195.5, prev_close: 195.4, change_percent: 1.72, volume: 32400000, amount: 6420000000, turnover_rate: 0.74, trade_date: "2026-05-29" },
};

let nextWatchId = 4;

export function addMockWatch(stockCode: string, groupName = "default", remark = ""): TWatchItem {
  const existing = mockQuotes[stockCode];
  const item: TWatchItem = {
    id: nextWatchId++,
    stock_code: stockCode,
    stock_name: existing?.stock_name ?? stockCode,
    group_name: groupName,
    remark,
  };
  mockWatchlist = [...mockWatchlist, item];
  if (!mockQuotes[stockCode]) {
    mockQuotes[stockCode] = {
      stock_code: stockCode,
      stock_name: stockCode,
      price: 10,
      open: 10,
      high: 10.2,
      low: 9.8,
      prev_close: 10,
      change_percent: 0,
      volume: 1000000,
      amount: 10000000,
      turnover_rate: 0.1,
      trade_date: "2026-05-29",
    };
  }
  return item;
}

export function removeMockWatch(id: number): void {
  mockWatchlist = mockWatchlist.filter((w) => w.id !== id);
}

export function getMockWatchlistQuotes(): TStockQuote[] {
  return mockWatchlist.map((w) => mockQuotes[w.stock_code]).filter(Boolean);
}

export function genMockKline(seed = 100, count = 120): TKline[] {
  const list: TKline[] = [];
  let price = seed;
  const start = new Date("2026-01-01");
  for (let i = 0; i < count; i++) {
    const open = price;
    const change = (Math.sin(i / 5) + (Math.random() - 0.5)) * 2;
    const close = Math.max(1, open + change);
    const high = Math.max(open, close) + Math.random();
    const low = Math.min(open, close) - Math.random();
    const date = new Date(start.getTime() + i * 86400000).toISOString().slice(0, 10);
    list.push({
      date,
      open: Number(open.toFixed(2)),
      high: Number(high.toFixed(2)),
      low: Number(low.toFixed(2)),
      close: Number(close.toFixed(2)),
      volume: Math.round(1000000 + Math.random() * 5000000),
      amount: Math.round(100000000 + Math.random() * 500000000),
    });
    price = close;
  }
  return list;
}
