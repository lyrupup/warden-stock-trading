import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { strategyApi } from "../api";
import type { TCreateStrategyParams, TIndicatorGroup } from "../types";

const KEY = ["strategy"] as const;

/** 策略列表 */
export function useStrategies(params?: { kw?: string; tag?: string }) {
  return useQuery({
    queryKey: [...KEY, "list", params ?? {}],
    queryFn: () => strategyApi.list(params),
  });
}

/** 策略详情 */
export function useStrategy(id: number | undefined) {
  return useQuery({
    queryKey: [...KEY, "detail", id],
    queryFn: () => strategyApi.detail(id as number),
    enabled: id != null,
  });
}

/** 新建策略 */
export function useCreateStrategy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: TCreateStrategyParams) => strategyApi.create(data),
    onSuccess: () => void qc.invalidateQueries({ queryKey: [...KEY, "list"] }),
  });
}

/** 更新策略 */
export function useUpdateStrategy(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: TCreateStrategyParams) => strategyApi.update(id, data),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: [...KEY, "detail", id] });
      void qc.invalidateQueries({ queryKey: [...KEY, "list"] });
    },
  });
}

/** 删除策略 */
export function useDeleteStrategy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => strategyApi.remove(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: [...KEY, "list"] }),
  });
}

/** 复制策略 */
export function useCopyStrategy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => strategyApi.copy(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: [...KEY, "list"] }),
  });
}

/** 更新指标定义 */
export function useUpdateIndicators(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (group: TIndicatorGroup) => strategyApi.updateIndicators(id, group),
    onSuccess: () => void qc.invalidateQueries({ queryKey: [...KEY, "detail", id] }),
  });
}

/** 保存 skill.md */
export function useSaveSkill(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (content: string) => strategyApi.saveSkill(id, content),
    onSuccess: () => void qc.invalidateQueries({ queryKey: [...KEY, "detail", id] }),
  });
}
