import { useQuery } from "@tanstack/react-query";
import { useUserConfigStore } from "@/stores/user-config-store";
import { marketApi } from "../api";
import { useIsTradingTime } from "./use-is-trading-time";

/**
 * 自选股实时行情（见 FRONTEND.md §6 M1）：
 * 交易时段内按 M7 配置频率（默认 5s）轮询，非交易时段停止。
 */
export function useWatchlistQuotes() {
  const refreshMs = useUserConfigStore((s) => s.marketRefreshMs);
  const inTrading = useIsTradingTime();
  return useQuery({
    queryKey: ["market", "watchlist-quotes"],
    queryFn: () => marketApi.watchlistQuotes(),
    refetchInterval: inTrading ? refreshMs : false,
  });
}
