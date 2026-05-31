/**
 * M2 选股策略 + 量化粗筛前端类型（对齐 openapi 与后端 strategy/factor/rule）。
 */

/** 比较操作符 */
export type TIndicatorOp = ">" | ">=" | "<" | "<=" | "==" | "!=";

/** 因子类型 */
export type TFactorType =
  | "ma"
  | "ma_align"
  | "bias"
  | "amplitude"
  | "amplitude_streak"
  | "pct_change"
  | "field"
  | "vol_ratio"
  | "const";

/** 因子表达式（规则左/右操作数） */
export type TFactorExpr = {
  type: TFactorType;
  period?: number;
  periods?: number[];
  direction?: "bull" | "bear";
  field?: string;
  threshold?: number;
  days?: number;
  eps?: number;
  value?: number | boolean;
};

/** 单条量化规则 */
export type TIndicatorRule = {
  left: TFactorExpr;
  op: TIndicatorOp;
  right: TFactorExpr;
};

/** 量化因子规则组（支持 and/or 与嵌套子组） */
export type TIndicatorGroup = {
  logic: "and" | "or";
  rules: TIndicatorRule[];
  groups?: TIndicatorGroup[];
};

/** 策略 */
export type TStrategy = {
  id: number;
  name: string;
  description: string;
  tags: string;
  status: number;
  indicators?: TIndicatorGroup | null;
  skill?: string;
  skill_version?: number;
  created_at: string;
};

/** 新建/更新策略参数 */
export type TCreateStrategyParams = {
  name: string;
  description?: string;
  tags?: string;
  indicators?: TIndicatorGroup | null;
};

/** 因子目录参数定义 */
export type TIndicatorParamSpec = {
  key: string;
  type: string;
  required: boolean;
  default?: unknown;
  enum?: unknown[];
  desc: string;
};

/** 因子目录项 */
export type TIndicatorCatalogItem = {
  type: TFactorType;
  name: string;
  value_type: "number" | "bool";
  unit?: string;
  description: string;
  params: TIndicatorParamSpec[];
};

/** 内置策略模板 */
export type TStrategyTemplate = {
  key: string;
  name: string;
  description: string;
  tags: string;
  scenario: string;
  indicators: TIndicatorGroup;
};

/** 股票池范围 */
export type TScreenUniverse = {
  type: "all" | "board" | "watchlist" | "codes";
  board?: string;
  codes?: string[];
};

/** 粗筛请求参数 */
export type TScreenParams = {
  universe: TScreenUniverse;
  trade_date?: string;
  limit?: number;
};

/** 粗筛候选 */
export type TScreenCandidate = {
  stock_code: string;
  stock_name: string;
  score: number;
  factors: Record<string, number>;
  matched: number[];
};

/** 粗筛任务状态：0 pending,1 running,2 done,3 failed */
export type TScreenStatus = 0 | 1 | 2 | 3;

/** 粗筛结果 */
export type TScreenResult = {
  id: number;
  strategy_id: number;
  status: TScreenStatus;
  trade_date?: string;
  universe_count: number;
  matched_count: number;
  candidates: TScreenCandidate[];
  error_msg?: string;
  created_at: string;
};
