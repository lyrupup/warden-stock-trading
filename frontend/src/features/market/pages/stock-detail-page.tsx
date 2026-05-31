import { useState } from "react";
import { useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/cn";
import { formatLargeNumber, formatPercent, formatPrice, getQuoteColor } from "@/lib/format";
import { KlineChart } from "../components/kline-chart";
import { useStockKline, useStockQuote } from "../hooks/use-stock-detail";
import type { TKlinePeriod } from "../types";

const PERIODS: { value: TKlinePeriod; label: string }[] = [
  { value: "day", label: "日" },
  { value: "week", label: "周" },
  { value: "month", label: "月" },
];

/** 个股详情 /market/:code（M1） */
export function StockDetailPage() {
  const { code = "" } = useParams();
  const { t } = useTranslation();
  const [period, setPeriod] = useState<TKlinePeriod>("day");

  const quoteQuery = useStockQuote(code);
  const klineQuery = useStockKline(code, period);
  const quote = quoteQuery.data;

  return (
    <div>
      <PageHeader
        title={quote ? `${quote.stock_name} ${quote.stock_code}` : code}
        description={t("market.title")}
      />

      {quoteQuery.isError ? (
        <EmptyState
          title={t("common.error")}
          actionLabel={t("common.retry")}
          onAction={() => void quoteQuery.refetch()}
        />
      ) : quote ? (
        <Card className="mb-6">
          <CardContent className="grid grid-cols-2 gap-4 p-6 sm:grid-cols-4">
            <Stat label={t("market.columns.price")} value={formatPrice(quote.price)} color={getQuoteColor(quote.change_percent)} />
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
