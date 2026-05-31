import dayjs from "dayjs";
import "dayjs/locale/zh-cn";
import customParseFormat from "dayjs/plugin/customParseFormat";
import isSameOrAfter from "dayjs/plugin/isSameOrAfter";
import isSameOrBefore from "dayjs/plugin/isSameOrBefore";
import relativeTime from "dayjs/plugin/relativeTime";

dayjs.extend(customParseFormat);
dayjs.extend(relativeTime);
dayjs.extend(isSameOrAfter);
dayjs.extend(isSameOrBefore);
dayjs.locale("zh-cn");

export { dayjs };

/** 格式化为日期时间（YYYY-MM-DD HH:mm:ss） */
export function formatDateTime(value?: string | number | Date | null): string {
  if (value === null || value === undefined || value === "") return "--";
  const d = dayjs(value);
  return d.isValid() ? d.format("YYYY-MM-DD HH:mm:ss") : "--";
}

/** 格式化为日期（YYYY-MM-DD） */
export function formatDate(value?: string | number | Date | null): string {
  if (value === null || value === undefined || value === "") return "--";
  const d = dayjs(value);
  return d.isValid() ? d.format("YYYY-MM-DD") : "--";
}

/** 相对时间（如「3 分钟前」） */
export function fromNow(value?: string | number | Date | null): string {
  if (value === null || value === undefined || value === "") return "--";
  const d = dayjs(value);
  return d.isValid() ? d.fromNow() : "--";
}
