import { useQuery } from "@tanstack/react-query";
import { useUserConfigStore } from "@/stores/user-config-store";
import { marketApi } from "../api";
import { useIsTradingTime } from "./use-is-trading-time";

/** 大盘指数：进入自动加载，交易时段内按配置频率轮询 */
export function useIndices() {
  const refreshMs = useUserConfigStore((s) => s.marketRefreshMs);
  const inTrading = useIsTradingTime();
  return useQuery({
    queryKey: ["market", "indices"],
    queryFn: () => marketApi.indices(),
    refetchInterval: inTrading ? refreshMs : false,
  });
}
