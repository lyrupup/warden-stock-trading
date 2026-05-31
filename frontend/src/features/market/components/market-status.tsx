import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Badge, type TBadgeProps } from "@/components/ui/badge";
import { dayjs } from "@/lib/date";
import { getTradingPhase } from "@/lib/trading-time";

/** 各交易阶段对应的 Badge 样式 */
const PHASE_VARIANT: Record<ReturnType<typeof getTradingPhase>, TBadgeProps["variant"]> = {
  trading: "success",
  preOpen: "default",
  lunchBreak: "secondary",
  closed: "secondary",
  nonTradingDay: "outline",
};

/**
 * 市场状态条：展示当天日期（含星期）+ 当前交易阶段，
 * 供行情中心 `/market` 与个股行情 `/market/quote` 页头复用。
 * 每分钟刷新一次，使盘中/午休/收盘切换无需手动刷新页面。
 */
export function MarketStatus() {
  const { t } = useTranslation();
  const [now, setNow] = useState(() => new Date());

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 60_000);
    return () => clearInterval(timer);
  }, []);

  const phase = getTradingPhase(now);

  return (
    <div className="flex items-center gap-2 text-sm">
      <span className="text-muted-foreground tabular-nums">
        {dayjs(now).format("YYYY-MM-DD ddd")}
      </span>
      <Badge variant={PHASE_VARIANT[phase]}>{t(`market.phase.${phase}`)}</Badge>
    </div>
  );
}
