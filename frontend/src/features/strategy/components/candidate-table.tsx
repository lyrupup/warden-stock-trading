import { useMemo } from "react";
import { DataTable, type TDataTableColumn } from "@/components/common/data-table";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/cn";
import { formatPrice } from "@/lib/format";
import type { TScreenCandidate } from "../types";

type TCandidateTableProps = {
  candidates: TScreenCandidate[];
  loading?: boolean;
  onAddWatch?: (code: string) => void;
  onAnalyze?: (code: string) => void;
};

/**
 * 因子展示类型：
 * - bool：布尔因子（ma_align / amplitude_streak），后端用 1/0 表达，前端渲染成 ✓ / ✗。
 * - percent：已乘 100 的百分比因子（bias / amplitude / pct_change），加 % 后缀。
 * - ratio：倍数因子（vol_ratio），保留小数不加单位。
 * - number：价格 / 均线 / 其他数值，按 formatPrice 渲染。
 */
type TFactorKind = "bool" | "percent" | "ratio" | "number";

type TFactorMeta = { label: string; kind: TFactorKind };

/** 固定键名的因子元数据。 */
const FACTOR_META: Record<string, TFactorMeta> = {
  ma5: { label: "MA5", kind: "number" },
  ma10: { label: "MA10", kind: "number" },
  ma20: { label: "MA20", kind: "number" },
  ma30: { label: "MA30", kind: "number" },
  ma60: { label: "MA60", kind: "number" },
  bias5: { label: "乖离5", kind: "percent" },
  bias10: { label: "乖离10", kind: "percent" },
  bias20: { label: "乖离20", kind: "percent" },
  amplitude: { label: "振幅", kind: "percent" },
  amplitude_streak: { label: "连续振幅", kind: "bool" },
  close: { label: "收盘", kind: "number" },
};

/** 由因子 key 派生标签 + 展示类型（动态前缀 key 走规则匹配）。导出用于单测。 */
export function factorMeta(key: string): TFactorMeta {
  const fixed = FACTOR_META[key];
  if (fixed) return fixed;
  if (key.startsWith("ma_align")) return { label: "排列", kind: "bool" };
  if (key.startsWith("pct_change")) return { label: "涨跌幅", kind: "percent" };
  if (key.startsWith("vol_ratio")) return { label: "量比", kind: "ratio" };
  return { label: key, kind: "number" };
}

/** 按展示类型渲染因子单元格。布尔因子用 ✓ / ✗ 上色，百分比因子加 % 后缀。 */
function renderFactor(value: number | null | undefined, kind: TFactorKind) {
  if (value == null || Number.isNaN(value)) {
    return <span className="text-muted-foreground">--</span>;
  }
  if (kind === "bool") {
    const ok = value >= 0.5;
    return (
      <span
        aria-label={ok ? "成立" : "不成立"}
        className={cn("inline-block tabular-nums", ok ? "text-quote-up" : "text-muted-foreground")}
      >
        {ok ? "✓" : "✗"}
      </span>
    );
  }
  if (kind === "percent") {
    return <span className="tabular-nums">{formatPrice(value, 2)}%</span>;
  }
  return <span className="tabular-nums">{formatPrice(value, 2)}</span>;
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
    ...factorKeys.map<TDataTableColumn<TScreenCandidate>>((key) => {
      const meta = factorMeta(key);
      return {
        key: `f_${key}`,
        header: meta.label,
        // 布尔列居中更直观，数值列右对齐对小数点。
        align: meta.kind === "bool" ? "center" : "right",
        // 排序统一按原始数值，避免被显示文本（✓/✗、% 后缀）影响。
        accessor: (r) => r.factors[key] ?? null,
        sortable: true,
        render: (r) => renderFactor(r.factors[key], meta.kind),
      };
    }),
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
