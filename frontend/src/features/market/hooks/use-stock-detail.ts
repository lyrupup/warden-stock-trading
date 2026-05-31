import { useQuery } from "@tanstack/react-query";
import { useUserConfigStore } from "@/stores/user-config-store";
import { marketApi } from "../api";
import type { TKlinePeriod } from "../types";
import { useIsTradingTime } from "./use-is-trading-time";

/** 个股行情详情 */
export function useStockQuote(code: string) {
  const refreshMs = useUserConfigStore((s) => s.marketRefreshMs);
  const inTrading = useIsTradingTime();
  return useQuery({
    queryKey: ["market", "stock", code],
    queryFn: () => marketApi.stock(code),
    enabled: Boolean(code),
    refetchInterval: inTrading ? refreshMs : false,
  });
}

/** 个股 K 线 */
export function useStockKline(code: string, period: TKlinePeriod = "day") {
  return useQuery({
    queryKey: ["market", "kline", code, period],
    queryFn: () => marketApi.kline(code, period),
    enabled: Boolean(code),
  });
}
