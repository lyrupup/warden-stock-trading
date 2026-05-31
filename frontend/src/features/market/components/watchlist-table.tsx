import { useTranslation } from "react-i18next";
import { DataTable, type TDataTableColumn } from "@/components/common/data-table";
import { QuoteCell } from "@/components/common/quote-cell";
import { Button } from "@/components/ui/button";
import { formatChange, formatLargeNumber, formatPercent, formatPrice } from "@/lib/format";
import type { TStockQuote } from "../types";

type TWatchlistTableProps = {
  data: TStockQuote[];
  loading?: boolean;
  onRowClick?: (quote: TStockQuote) => void;
  onRemove?: (stockCode: string) => void;
};

/** 自选股表格：基于通用 data-table，涨跌色用 QuoteCell */
export function WatchlistTable({ data, loading, onRowClick, onRemove }: TWatchlistTableProps) {
  const { t } = useTranslation();

  const columns: TDataTableColumn<TStockQuote>[] = [
    {
      key: "stock_code",
      header: t("market.columns.code"),
      accessor: (r) => r.stock_code,
      sortable: true,
      render: (r) => <span className="font-mono text-xs">{r.stock_code}</span>,
    },
    {
      key: "stock_name",
      header: t("market.columns.name"),
      accessor: (r) => r.stock_name,
      render: (r) => <span className="font-medium">{r.stock_name}</span>,
    },
    {
      key: "price",
      header: t("market.columns.price"),
      align: "right",
      accessor: (r) => r.price,
      sortable: true,
      render: (r) => (
        <QuoteCell value={formatPrice(r.price)} change={r.change_percent} />
      ),
    },
    {
      key: "change_percent",
      header: t("market.columns.changePercent"),
      align: "right",
      accessor: (r) => r.change_percent,
      sortable: true,
      render: (r) => (
        <QuoteCell value={formatPercent(r.change_percent)} change={r.change_percent} />
      ),
    },
    {
      key: "change",
      header: t("market.columns.change"),
      align: "right",
      render: (r) => {
        const change = r.price - r.prev_close;
        return <QuoteCell value={formatChange(change)} change={r.change_percent} />;
      },
    },
    {
      key: "turnover_rate",
      header: t("market.columns.turnoverRate"),
      align: "right",
      accessor: (r) => r.turnover_rate,
      sortable: true,
      render: (r) => <span className="tabular-nums">{formatPercent(r.turnover_rate)}</span>,
    },
    {
      key: "amount",
      header: t("market.columns.amount"),
      align: "right",
      accessor: (r) => r.amount,
      sortable: true,
      render: (r) => <span className="tabular-nums">{formatLargeNumber(r.amount)}</span>,
    },
    ...(onRemove
      ? [
          {
            key: "actions",
            header: "",
            align: "right" as const,
            render: (r: TStockQuote) => (
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation();
                  onRemove(r.stock_code);
                }}
              >
                {t("common.delete")}
              </Button>
            ),
          },
        ]
      : []),
  ];

  return (
    <DataTable
      columns={columns}
      data={data}
      rowKey={(r) => r.stock_code}
      loading={loading}
      emptyText={t("common.empty")}
      onRowClick={onRowClick}
    />
  );
}
