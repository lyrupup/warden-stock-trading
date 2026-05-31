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
