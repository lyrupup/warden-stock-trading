/**
 * 金额 / 股价 / 涨跌色统一格式化工具。
 *
 * A 股习惯：涨 = 红、跌 = 绿。涨跌色统一通过 `getQuoteColor` 返回 Tailwind class，
 * 禁止在组件内散落处理颜色逻辑（见 FRONTEND.md §1.2 / §4.5）。
 */

/** 涨跌方向 */
export type TQuoteDirection = "up" | "down" | "flat";

/** 根据涨跌幅 / 涨跌额判断方向 */
export function getQuoteDirection(change?: number | null): TQuoteDirection {
  if (change === null || change === undefined || Number.isNaN(change) || change === 0) {
    return "flat";
  }
  return change > 0 ? "up" : "down";
}

/** 根据涨跌幅返回涨跌色的 Tailwind 文本 class（涨红跌绿） */
export function getQuoteColor(change?: number | null): string {
  switch (getQuoteDirection(change)) {
    case "up":
      return "text-quote-up";
    case "down":
      return "text-quote-down";
    default:
      return "text-quote-flat";
  }
}

/** 格式化价格 / 金额（保留固定小数位，带千分位） */
export function formatPrice(value?: number | null, fractionDigits = 2): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "--";
  return value.toLocaleString("zh-CN", {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  });
}

/** 格式化涨跌幅百分比（带正负号），如 +1.23% / -0.50% */
export function formatPercent(value?: number | null, fractionDigits = 2): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "--";
  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(fractionDigits)}%`;
}

/** 格式化涨跌额（带正负号） */
export function formatChange(value?: number | null, fractionDigits = 2): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "--";
  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(fractionDigits)}`;
}

/**
 * 大额数字简写：成交量 / 成交额按 万 / 亿 单位压缩。
 */
export function formatLargeNumber(value?: number | null, fractionDigits = 2): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "--";
  const abs = Math.abs(value);
  if (abs >= 1e8) return `${(value / 1e8).toFixed(fractionDigits)}亿`;
  if (abs >= 1e4) return `${(value / 1e4).toFixed(fractionDigits)}万`;
  return value.toFixed(0);
}
