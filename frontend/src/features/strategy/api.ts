import { http } from "@/core/http-client";
import { toNumber } from "@/lib/decimal";
import type {
  TCreateStrategyParams,
  TIndicatorCatalogItem,
  TIndicatorGroup,
  TScreenCandidate,
  TScreenParams,
  TScreenResult,
  TStrategy,
  TStrategyTemplate,
} from "./types";

/** 候选数值字段（score / factors）后端为 decimal 字符串，统一转 number。 */
function normalizeCandidate(c: TScreenCandidate): TScreenCandidate {
  const factors: Record<string, number> = {};
  for (const [k, v] of Object.entries(c.factors ?? {})) {
    factors[k] = toNumber(v as unknown as string);
  }
  return { ...c, score: toNumber(c.score as unknown as string), factors };
}

function normalizeScreenResult(r: TScreenResult): TScreenResult {
  return { ...r, candidates: (r.candidates ?? []).map(normalizeCandidate) };
}

/**
 * M2 策略 + 量化粗筛接口（对齐 openapi /strategies/*）。
 * 工厂函数注入 http，便于测试 Mock（见 FRONTEND.md §4.2）。
 */
export function createStrategyApi(client = http) {
  return {
    /** 策略列表 */
    list: (params?: { kw?: string; tag?: string }) =>
      client.get<TStrategy[]>("strategies", {
        searchParams: cleanParams(params),
      }),
    /** 策略详情（含指标、skill） */
    detail: (id: number) => client.get<TStrategy>(`strategies/${id}`),
    /** 新建策略 */
    create: (data: TCreateStrategyParams) => client.post<TStrategy>("strategies", { json: data }),
    /** 更新策略 */
    update: (id: number, data: TCreateStrategyParams) =>
      client.put<void>(`strategies/${id}`, { json: data }),
    /** 删除策略 */
    remove: (id: number) => client.delete<void>(`strategies/${id}`),
    /** 复制策略 */
    copy: (id: number) => client.post<TStrategy>(`strategies/${id}/copy`),
    /** 更新指标定义 */
    updateIndicators: (id: number, group: TIndicatorGroup) =>
      client.put<void>(`strategies/${id}/indicators`, { json: group }),
    /** 获取 skill.md */
    getSkill: (id: number) => client.get<{ content: string; version: number }>(`strategies/${id}/skill`),
    /** 保存 skill.md（生成新版本） */
    saveSkill: (id: number, content: string) =>
      client.put<void>(`strategies/${id}/skill`, { json: { content } }),
    /** 量化因子目录 */
    catalog: () => client.get<TIndicatorCatalogItem[]>("strategies/indicators/catalog"),
    /** 内置策略模板 */
    templates: () => client.get<TStrategyTemplate[]>("strategies/templates"),
    /** 发起量化粗筛（异步） */
    runScreen: (id: number, params: TScreenParams) =>
      client.post<{ taskId: string }>(`strategies/${id}/screen`, { json: params }),
    /** 查询粗筛结果 */
    screenResult: (id: number, taskId: string) =>
      client.get<TScreenResult>(`strategies/${id}/screen/${taskId}`).then(normalizeScreenResult),
    /** 最近一次粗筛结果 */
    screenLatest: (id: number) =>
      client.get<TScreenResult>(`strategies/${id}/screen/latest`).then(normalizeScreenResult),
    /** 同步快速粗筛预览 */
    previewScreen: (params: TScreenParams & { indicators: TIndicatorGroup }) =>
      client.post<TScreenResult>("strategies/screen/preview", { json: params }).then(normalizeScreenResult),
  };
}

function cleanParams(params?: Record<string, string | undefined>): Record<string, string> {
  const out: Record<string, string> = {};
  if (!params) return out;
  for (const [k, v] of Object.entries(params)) {
    if (v != null && v !== "") out[k] = v;
  }
  return out;
}

export const strategyApi = createStrategyApi();

export type TStrategyApi = ReturnType<typeof createStrategyApi>;
