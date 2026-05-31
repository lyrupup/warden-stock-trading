/**
 * decimal 统一解析工具。
 *
 * 后端金额/价格/数量/比率字段用 shopspring/decimal，JSON 序列化为**字符串**
 * （如 "10.5000"，见 openapi components.schemas.Decimal）。前端务必在数据入口
 * 经此工具转成 number 后再计算 / 格式化，禁止直接对其做算术或 `.toFixed()`。
 */

/** 可被解析的 decimal 输入：兼容 number 与后端 decimal 序列化出的字符串。 */
export type TNumericInput = number | string | null | undefined;

/**
 * 统一数值强转：兼容 number / 字符串。无法解析时返回 NaN，交由调用方兜底。
 */
export function toNumber(value: TNumericInput): number {
  if (typeof value === "number") return value;
  if (value === null || value === undefined || value === "") return NaN;
  return Number(value);
}

/**
 * 把对象中指定的 decimal 字符串字段批量转成 number，返回浅拷贝。
 * 用于各 feature 的 API 边界归一化（M1 行情、M3 持仓、M2 回测等复用）。
 *
 * @example
 * const pos = coerceDecimalFields(raw, ["quantity", "avg_cost", "float_pnl"]);
 */
export function coerceDecimalFields<T extends Record<string, unknown>>(
  obj: T,
  keys: readonly (keyof T)[],
): T {
  const out = { ...obj };
  for (const key of keys) {
    if (key in out) {
      out[key] = toNumber(out[key] as TNumericInput) as T[keyof T];
    }
  }
  return out;
}
