import { useState } from "react";
import { useTranslation } from "react-i18next";
import { EmptyState } from "@/components/common/empty-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/cn";
import { formatLargeNumber, formatPercent, formatPrice, getQuoteColor } from "@/lib/format";
import { useStockKline, useStockQuote } from "../hooks/use-stock-detail";
import type { TKlinePeriod } from "../types";
import { KlineChart } from "./kline-chart";

const PERIODS: { value: TKlinePeriod; label: string }[] = [
  { value: "day", label: "日" },
  { value: "week", label: "周" },
  { value: "month", label: "月" },
];

type TStockQuotePanelProps = {
  code: string;
};

/**
 * 个股行情面板：当日行情卡片 + K 线图（M1）。
 * 自包含数据获取（react-query 按 code 缓存），供「个股详情页」与「行情中心·个股行情 tab」复用。
 */
export function StockQuotePanel({ code }: TStockQuotePanelProps) {
  const { t } = useTranslation();
  const [period, setPeriod] = useState<TKlinePeriod>("day");

  const quoteQuery = useStockQuote(code);
  const klineQuery = useStockKline(code, period);
  const quote = quoteQuery.data;

  return (
    <div>
      {quoteQuery.isError ? (
        <EmptyState
          title={t("common.error")}
          actionLabel={t("common.retry")}
          onAction={() => void quoteQuery.refetch()}
        />
      ) : quote ? (
        <Card className="mb-6">
          <CardContent className="grid grid-cols-2 gap-4 p-6 sm:grid-cols-4">
            <Stat
              label={t("market.columns.price")}
              value={formatPrice(quote.price)}
              color={getQuoteColor(quote.change_percent)}
            />
            <Stat
              label={t("market.columns.changePercent")}
              value={formatPercent(quote.change_percent)}
              color={getQuoteColor(quote.change_percent)}
            />
            <Stat label={t("market.columns.open")} value={formatPrice(quote.open)} />
            <Stat label={t("market.columns.high")} value={formatPrice(quote.high)} />
            <Stat label={t("market.columns.low")} value={formatPrice(quote.low)} />
            <Stat label={t("market.columns.turnoverRate")} value={formatPercent(quote.turnover_rate)} />
            <Stat label={t("market.columns.volume")} value={formatLargeNumber(quote.volume)} />
            <Stat label={t("market.columns.amount")} value={formatLargeNumber(quote.amount)} />
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardContent className="p-6">
          <div className="mb-4 flex items-center gap-2">
            {PERIODS.map((p) => (
              <Button
                key={p.value}
                variant={period === p.value ? "default" : "outline"}
                size="sm"
                onClick={() => setPeriod(p.value)}
              >
                {p.label}
              </Button>
            ))}
          </div>
          {klineQuery.data && klineQuery.data.length > 0 ? (
            <KlineChart data={klineQuery.data} />
          ) : (
            <EmptyState title={t("common.empty")} />
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function Stat({ label, value, color }: { label: string; value: string; color?: string }) {
  return (
    <div>
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className={cn("mt-1 text-lg font-semibold tabular-nums", color)}>{value}</div>
    </div>
  );
}
