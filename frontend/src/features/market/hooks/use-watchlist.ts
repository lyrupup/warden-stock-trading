import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { marketApi } from "../api";
import type { TAddWatchParams } from "../types";

/** 自选股列表（基础信息，不含实时行情） */
export function useWatchlist() {
  return useQuery({
    queryKey: ["market", "watchlist"],
    queryFn: () => marketApi.watchlist(),
  });
}

/** 添加自选股，成功后失效自选相关查询 */
export function useAddWatch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: TAddWatchParams) => marketApi.addWatch(data),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["market", "watchlist"] });
      void queryClient.invalidateQueries({ queryKey: ["market", "watchlist-quotes"] });
    },
  });
}

/** 删除自选股 */
export function useRemoveWatch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => marketApi.removeWatch(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["market", "watchlist"] });
      void queryClient.invalidateQueries({ queryKey: ["market", "watchlist-quotes"] });
    },
  });
}
