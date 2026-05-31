import type { TFactorExpr, TFactorType, TIndicatorCatalogItem, TIndicatorOp } from "../types";

/** 比较操作符与中文标签 */
export const OPS: { value: TIndicatorOp; label: string }[] = [
  { value: ">", label: "大于 >" },
  { value: ">=", label: "大于等于 ≥" },
  { value: "<", label: "小于 <" },
  { value: "<=", label: "小于等于 ≤" },
  { value: "==", label: "等于 ==" },
  { value: "!=", label: "不等于 !=" },
];

/** 依据因子目录构造某类型因子的默认表达式（带默认参数）。 */
export function defaultFactor(type: TFactorType, catalog: TIndicatorCatalogItem[]): TFactorExpr {
  if (type === "const") return { type: "const", value: 0 };
  const item = catalog.find((c) => c.type === type);
  const expr: TFactorExpr = { type };
  if (!item) return expr;
  for (const p of item.params) {
    applyParam(expr, p.key, p.default);
  }
  return expr;
}

/** 把单个参数值写入因子表达式（统一在此约束 key → 字段映射）。 */
export function applyParam(expr: TFactorExpr, key: string, value: unknown): void {
  switch (key) {
    case "period":
      expr.period = toInt(value, 5);
      break;
    case "periods":
      expr.periods = Array.isArray(value) ? value.map((v) => toInt(v, 0)) : [5, 10, 20];
      break;
    case "direction":
      expr.direction = value === "bear" ? "bear" : "bull";
      break;
    case "field":
      expr.field = typeof value === "string" ? value : "close";
      break;
    case "threshold":
      expr.threshold = toNum(value, 5);
      break;
    case "days":
      expr.days = toInt(value, 3);
      break;
    case "eps":
      expr.eps = toNum(value, 0);
      break;
    default:
      break;
  }
}

/** 因子的人类可读描述（用于规则摘要/表头）。 */
export function describeFactor(expr: TFactorExpr): string {
  switch (expr.type) {
    case "ma":
      return `MA${expr.period ?? ""}`;
    case "ma_align": {
      const ps = (expr.periods ?? []).join(expr.direction === "bear" ? "<" : ">");
      return `${ps}（${expr.direction === "bear" ? "空头" : "多头"}排列）`;
    }
    case "bias":
      return `乖离率(${expr.period ?? ""})`;
    case "amplitude":
      return "振幅";
    case "amplitude_streak":
      return `连续振幅≥${expr.threshold ?? ""}%×${expr.days ?? ""}日`;
    case "pct_change":
      return `${expr.period ?? ""}日涨跌幅`;
    case "vol_ratio":
      return `量比(${expr.period ?? ""})`;
    case "field":
      return fieldLabel(expr.field);
    case "const":
      return typeof expr.value === "boolean" ? (expr.value ? "真" : "假") : `常量 ${expr.value ?? 0}`;
    default:
      return expr.type;
  }
}

const FIELD_LABELS: Record<string, string> = {
  close: "收盘价",
  open: "开盘价",
  high: "最高价",
  low: "最低价",
  prev_close: "前收盘",
  volume: "成交量",
  amount: "成交额",
  change_percent: "涨跌幅",
};

export function fieldLabel(field?: string): string {
  return field ? (FIELD_LABELS[field] ?? field) : "字段";
}

/**
 * 因子参数 key → 构造器中显示在输入框前的小标签。
 * 后端因子目录返回的 `desc` 是较长的说明文案（用作 tooltip/aria），
 * 这里给一个精炼标签，避免输入框旁边一片空白让用户不知道在填什么。
 */
const PARAM_LABELS: Record<string, string> = {
  period: "周期",
  periods: "周期组",
  direction: "方向",
  field: "字段",
  threshold: "阈值%",
  days: "天数",
  eps: "容差",
};

export function paramLabel(key: string): string {
  return PARAM_LABELS[key] ?? key;
}

/** 排列方向枚举 → 业务友好文案。后端 enum 仍是 bull/bear，前端做翻译。 */
const DIRECTION_LABELS: Record<string, string> = {
  bull: "多头（严格递减）",
  bear: "空头（严格递增）",
};

export function directionLabel(v: string): string {
  return DIRECTION_LABELS[v] ?? v;
}

/** 因子是否产出布尔值（布尔因子默认与 const(true) 比较）。 */
export function isBoolFactor(type: TFactorType): boolean {
  return type === "ma_align" || type === "amplitude_streak";
}

function toInt(v: unknown, fallback: number): number {
  const n = Number(v);
  return Number.isFinite(n) ? Math.trunc(n) : fallback;
}

function toNum(v: unknown, fallback: number): number {
  const n = Number(v);
  return Number.isFinite(n) ? n : fallback;
}
