import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { TStockQuote } from "../types";
import { WatchlistTable } from "./watchlist-table";

const quotes: TStockQuote[] = [
  {
    stock_code: "600519",
    stock_name: "贵州茅台",
    price: 1689.5,
    open: 1672,
    high: 1695,
    low: 1668,
    prev_close: 1675,
    change_percent: 0.87,
    volume: 2150000,
    amount: 3630000000,
    turnover_rate: 0.17,
    trade_date: "2026-05-29",
  },
  {
    stock_code: "000858",
    stock_name: "五粮液",
    price: 142.3,
    open: 144,
    high: 144.5,
    low: 141.8,
    prev_close: 143.8,
    change_percent: -1.04,
    volume: 18500000,
    amount: 2640000000,
    turnover_rate: 0.48,
    trade_date: "2026-05-29",
  },
];

describe("WatchlistTable", () => {
  it("渲染自选股名称与格式化后的现价 / 涨跌幅", () => {
    render(<WatchlistTable data={quotes} />);
    expect(screen.getByText("贵州茅台")).toBeInTheDocument();
    expect(screen.getByText("1,689.50")).toBeInTheDocument();
    expect(screen.getByText("+0.87%")).toBeInTheDocument();
    expect(screen.getByText("-1.04%")).toBeInTheDocument();
  });

  it("涨用红色 class、跌用绿色 class（A 股涨红跌绿）", () => {
    render(<WatchlistTable data={quotes} />);
    expect(screen.getByText("+0.87%")).toHaveClass("text-quote-up");
    expect(screen.getByText("-1.04%")).toHaveClass("text-quote-down");
  });

  it("点击删除触发 onRemove 回调并传入股票代码", async () => {
    const onRemove = vi.fn();
    render(<WatchlistTable data={quotes} onRemove={onRemove} />);
    const buttons = screen.getAllByRole("button", { name: "删除" });
    await userEvent.click(buttons[0]);
    expect(onRemove).toHaveBeenCalledWith("600519");
  });
});
