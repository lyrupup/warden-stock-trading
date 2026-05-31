import { useEffect, useState } from "react";
import { isTradingTime } from "@/lib/trading-time";

/**
 * 响应式交易时段判断：每 30s 重新计算一次，用于驱动行情轮询开关。
 */
export function useIsTradingTime(): boolean {
  const [inTrading, setInTrading] = useState(() => isTradingTime());

  useEffect(() => {
    const timer = setInterval(() => setInTrading(isTradingTime()), 30_000);
    return () => clearInterval(timer);
  }, []);

  return inTrading;
}
