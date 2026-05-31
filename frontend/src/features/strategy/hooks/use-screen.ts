import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { strategyApi } from "../api";
import type { TIndicatorGroup, TScreenParams } from "../types";

/**
 * 量化粗筛：发起任务 → 轮询结果直至 done/failed。
 * V1 后端为请求内同步执行，故任务返回时通常已完成；仍保留轮询以兼容后续异步化。
 */
export function useRunScreen(strategyId: number) {
  const [taskId, setTaskId] = useState<string | null>(null);

  const run = useMutation({
    mutationFn: (params: TScreenParams) => strategyApi.runScreen(strategyId, params),
    onSuccess: (data) => setTaskId(data.taskId),
  });

  const result = useQuery({
    queryKey: ["strategy", "screen", strategyId, taskId],
    queryFn: () => strategyApi.screenResult(strategyId, taskId as string),
    enabled: taskId != null,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status === 2 || status === 3 ? false : 1500;
    },
  });

  return { run, result, taskId };
}

/** 同步快速粗筛预览（携带临时指标） */
export function usePreviewScreen() {
  return useMutation({
    mutationFn: (params: TScreenParams & { indicators: TIndicatorGroup }) =>
      strategyApi.previewScreen(params),
  });
}
