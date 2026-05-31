import { dayjs } from "./date";

/**
 * A 股交易时段判断（用于控制行情轮询，避免非交易时段无效请求）。
 *
 * 交易时段（北京时间）：周一至周五 09:30-11:30、13:00-15:00。
 * 注意：未接入交易日历，节假日不在此判断（由后端 / 交易日历进一步裁剪）。
 */

const MORNING_START = 9 * 60 + 30;
const MORNING_END = 11 * 60 + 30;
const AFTERNOON_START = 13 * 60;
const AFTERNOON_END = 15 * 60;

/** 是否为交易日（周一至周五，不含节假日） */
export function isTradingDay(date: Date = new Date()): boolean {
  const day = dayjs(date).day();
  return day >= 1 && day <= 5;
}

/** 当前（或指定时间）是否处于交易时段 */
export function isTradingTime(date: Date = new Date()): boolean {
  if (!isTradingDay(date)) return false;
  const d = dayjs(date);
  const minutes = d.hour() * 60 + d.minute();
  const inMorning = minutes >= MORNING_START && minutes <= MORNING_END;
  const inAfternoon = minutes >= AFTERNOON_START && minutes <= AFTERNOON_END;
  return inMorning || inAfternoon;
}

/**
 * 交易阶段，用于页面直观展示当前盘口状态：
 * - nonTradingDay 非交易日（周末，节假日由后端裁剪）
 * - preOpen 未开盘（交易日 09:30 前）
 * - trading 交易中（09:30-11:30 / 13:00-15:00）
 * - lunchBreak 午间休市（11:30-13:00）
 * - closed 已收盘（交易日 15:00 后）
 */
export type TTradingPhase = "nonTradingDay" | "preOpen" | "trading" | "lunchBreak" | "closed";

/** 判断当前（或指定时间）所处的交易阶段 */
export function getTradingPhase(date: Date = new Date()): TTradingPhase {
  if (!isTradingDay(date)) return "nonTradingDay";
  const d = dayjs(date);
  const minutes = d.hour() * 60 + d.minute();
  if (minutes < MORNING_START) return "preOpen";
  if (minutes <= MORNING_END) return "trading";
  if (minutes < AFTERNOON_START) return "lunchBreak";
  if (minutes <= AFTERNOON_END) return "trading";
  return "closed";
}
