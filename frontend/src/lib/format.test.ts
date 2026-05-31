import { describe, expect, it } from "vitest";
import {
  formatChange,
  formatLargeNumber,
  formatPercent,
  formatPrice,
  getQuoteColor,
  getQuoteDirection,
} from "./format";

describe("getQuoteDirection", () => {
  it("正数为涨、负数为跌、0 与空为平", () => {
    expect(getQuoteDirection(1.2)).toBe("up");
    expect(getQuoteDirection(-0.5)).toBe("down");
    expect(getQuoteDirection(0)).toBe("flat");
    expect(getQuoteDirection(null)).toBe("flat");
    expect(getQuoteDirection(undefined)).toBe("flat");
  });
});

describe("getQuoteColor（A 股涨红跌绿）", () => {
  it("涨返回红色 class，跌返回绿色 class，平返回灰色 class", () => {
    expect(getQuoteColor(1)).toBe("text-quote-up");
    expect(getQuoteColor(-1)).toBe("text-quote-down");
    expect(getQuoteColor(0)).toBe("text-quote-flat");
  });
});

describe("formatPrice", () => {
  it("保留两位小数并加千分位", () => {
    expect(formatPrice(1689.5)).toBe("1,689.50");
    expect(formatPrice(0)).toBe("0.00");
  });
  it("非法值返回 --", () => {
    expect(formatPrice(null)).toBe("--");
    expect(formatPrice(undefined)).toBe("--");
    expect(formatPrice(Number.NaN)).toBe("--");
  });
});

describe("formatPercent / formatChange", () => {
  it("带正负号", () => {
    expect(formatPercent(1.234)).toBe("+1.23%");
    expect(formatPercent(-0.5)).toBe("-0.50%");
    expect(formatChange(2.5)).toBe("+2.50");
    expect(formatChange(-2.5)).toBe("-2.50");
  });
});

describe("formatLargeNumber", () => {
  it("按万 / 亿压缩", () => {
    expect(formatLargeNumber(12000)).toBe("1.20万");
    expect(formatLargeNumber(345600000000)).toBe("3456.00亿");
    expect(formatLargeNumber(500)).toBe("500");
  });
});
