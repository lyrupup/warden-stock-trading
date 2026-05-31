import { describe, expect, it } from "vitest";
import { isTradingDay, isTradingTime } from "./trading-time";

// 使用本地时间字符串（不带时区偏移），保证 hour()/day() 与测试机时区无关
describe("isTradingDay", () => {
  it("周一至周五为交易日，周末不是", () => {
    expect(isTradingDay(new Date("2026-05-25T10:00:00"))).toBe(true); // 周一
    expect(isTradingDay(new Date("2026-05-30T10:00:00"))).toBe(false); // 周六
    expect(isTradingDay(new Date("2026-05-31T10:00:00"))).toBe(false); // 周日
  });
});

describe("isTradingTime", () => {
  it("交易日 10:00 / 14:00 处于交易时段", () => {
    expect(isTradingTime(new Date("2026-05-25T10:00:00"))).toBe(true);
    expect(isTradingTime(new Date("2026-05-25T14:00:00"))).toBe(true);
  });
  it("交易日午休 12:00 与盘后 16:00 不在交易时段", () => {
    expect(isTradingTime(new Date("2026-05-25T12:00:00"))).toBe(false);
    expect(isTradingTime(new Date("2026-05-25T16:00:00"))).toBe(false);
  });
  it("非交易日任何时间都不在交易时段", () => {
    expect(isTradingTime(new Date("2026-05-30T10:00:00"))).toBe(false);
  });
});
