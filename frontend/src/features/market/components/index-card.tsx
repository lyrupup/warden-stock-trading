import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/cn";
import { formatChange, formatPercent, formatPrice, getQuoteColor } from "@/lib/format";
import type { TIndexQuote } from "../types";

type TIndexCardProps = {
  quote: TIndexQuote;
};

/** 大盘指数卡片：现价 + 涨跌额 + 涨跌幅，涨红跌绿 */
export function IndexCard({ quote }: TIndexCardProps) {
  const color = getQuoteColor(quote.change_percent);
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-baseline justify-between">
          <span className="text-sm font-medium text-muted-foreground">{quote.index_name}</span>
          <span className="text-xs text-muted-foreground">{quote.index_code}</span>
        </div>
        <div className={cn("mt-2 text-2xl font-semibold tabular-nums", color)}>
          {formatPrice(quote.price)}
        </div>
        <div className={cn("mt-1 flex items-center gap-3 text-sm tabular-nums", color)}>
          <span>{formatChange(quote.change_amount)}</span>
          <span>{formatPercent(quote.change_percent)}</span>
        </div>
      </CardContent>
    </Card>
  );
}
