import { cn } from "@/lib/cn";
import { getQuoteColor } from "@/lib/format";

type TQuoteCellProps = {
  /** 显示文本（已格式化），如 "+1.23%" 或 "12.34" */
  value: string;
  /** 用于决定涨跌色的数值（正涨负跌） */
  change?: number | null;
  className?: string;
};

/**
 * 行情涨跌色单元格（见 FRONTEND.md §2 common/quote-cell）：
 * 统一通过 getQuoteColor 渲染涨红跌绿，禁止散落颜色逻辑。
 */
export function QuoteCell({ value, change, className }: TQuoteCellProps) {
  return (
    <span className={cn("tabular-nums font-medium", getQuoteColor(change), className)}>{value}</span>
  );
}
