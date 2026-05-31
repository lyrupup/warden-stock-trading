import { describe, expect, it } from "vitest";
import { computeMA } from "./kline-chart";
import type { TKline } from "../types";

function mk(closes: number[]): TKline[] {
  return closes.map((c, i) => ({
    date: `2026-05-${String(i + 1).padStart(2, "0")}`,
    open: c,
    high: c,
    low: c,
    close: c,
    volume: 0,
    amount: 0,
  }));
}

describe("computeMA", () => {
  it("样本不足 period 返回空", () => {
    expect(computeMA(mk([1, 2, 3]), 5)).toEqual([]);
  });

  it("窗口对齐：MA3 在第 3 天起每天输出三日平均", () => {
    const out = computeMA(mk([1, 2, 3, 4, 5]), 3);
    expect(out).toEqual([
      { time: "2026-05-03", value: 2 }, // (1+2+3)/3
      { time: "2026-05-04", value: 3 }, // (2+3+4)/3
      { time: "2026-05-05", value: 4 }, // (3+4+5)/3
    ]);
  });

  it("period<=0 返回空，避免除零", () => {
    expect(computeMA(mk([1, 2, 3]), 0)).toEqual([]);
  });
});
