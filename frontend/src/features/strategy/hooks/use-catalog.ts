import { useQuery } from "@tanstack/react-query";
import { strategyApi } from "../api";

/** 量化因子目录（长缓存，驱动指标构造器渲染） */
export function useIndicatorCatalog() {
  return useQuery({
    queryKey: ["strategy", "catalog"],
    queryFn: () => strategyApi.catalog(),
    staleTime: 60 * 60 * 1000,
  });
}

/** 内置策略模板 */
export function useStrategyTemplates() {
  return useQuery({
    queryKey: ["strategy", "templates"],
    queryFn: () => strategyApi.templates(),
    staleTime: 60 * 60 * 1000,
  });
}
