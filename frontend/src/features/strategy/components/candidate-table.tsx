import { useMemo } from "react";
import { DataTable, type TDataTableColumn } from "@/components/common/data-table";
import { Button } from "@/components/ui/button";
import { formatPrice } from "@/lib/format";
import type { TScreenCandidate } from "../types";

type TCandidateTableProps = {
  candidates: TScreenCandidate[];
  loading?: boolean;
  onAddWatch?: (code: string) => void;
  onAnalyze?: (code: string) => void;
};

/** 因子键 → 中文表头（未命中的键回退原始 key）。 */
const FACTOR_LABELS: Record<string, string> = {
  ma5: "MA5",
  ma10: "MA10",
  ma20: "MA20",
  ma30: "MA30",
  ma60: "MA60",
  bias5: "乖离5",
  bias10: "乖离10",
  bias20: "乖离20",
  amplitude: "振幅",
  close: "收盘",
};

function factorHeader(key: string): string {
  if (FACTOR_LABELS[key]) return FACTOR_LABELS[key];
  if (key.startsWith("ma_align")) return "排列";
  if (key.startsWith("pct_change")) return "涨跌幅";
  if (key.startsWith("vol_ratio")) return "量比";
  return key;
}

/** 粗筛候选结果表：动态因子列 + 评分 + 操作（加自选 / AI 分析）。 */
export function CandidateTable({ candidates, loading, onAddWatch, onAnalyze }: TCandidateTableProps) {
  // 汇总所有候选出现过的因子键，作为动态列（保持稳定顺序）。
  const factorKeys = useMemo(() => {
    const set = new Set<string>();
    for (const c of candidates) {
      for (const k of Object.keys(c.factors ?? {})) set.add(k);
    }
    return Array.from(set).sort();
  }, [candidates]);

  const columns: TDataTableColumn<TScreenCandidate>[] = [
    {
      key: "stock_code",
      header: "代码",
      accessor: (r) => r.stock_code,
      sortable: true,
      render: (r) => <span className="font-mono text-xs">{r.stock_code}</span>,
    },
    {
      key: "stock_name",
      header: "名称",
      render: (r) => <span className="font-medium">{r.stock_name || "--"}</span>,
    },
    {
      key: "score",
      header: "评分",
      align: "right",
      accessor: (r) => r.score,
      sortable: true,
      render: (r) => <span className="tabular-nums">{r.score.toFixed(2)}</span>,
    },
    ...factorKeys.map<TDataTableColumn<TScreenCandidate>>((key) => ({
      key: `f_${key}`,
      header: factorHeader(key),
      align: "right",
      accessor: (r) => r.factors[key] ?? null,
      sortable: true,
      render: (r) => {
        const v = r.factors[key];
        return <span className="tabular-nums">{v == null ? "--" : formatPrice(v, 2)}</span>;
      },
    })),
    {
      key: "actions",
      header: "",
      align: "right",
      render: (r) => (
        <div className="flex justify-end gap-1">
          {onAddWatch ? (
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                onAddWatch(r.stock_code);
              }}
            >
              加自选
            </Button>
          ) : null}
          {onAnalyze ? (
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                onAnalyze(r.stock_code);
              }}
            >
              AI 分析
            </Button>
          ) : null}
        </div>
      ),
    },
  ];

  return (
    <DataTable
      columns={columns}
      data={candidates}
      rowKey={(r) => r.stock_code}
      loading={loading}
      emptyText="无符合条件的候选股"
    />
  );
}
