import { useMemo, useState, type ReactNode } from "react";
import { cn } from "@/lib/cn";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EmptyState } from "@/components/common/empty-state";

export type TColumnAlign = "left" | "center" | "right";

/** 列配置 */
export type TDataTableColumn<T> = {
  /** 列唯一 key */
  key: string;
  /** 表头 */
  header: ReactNode;
  /** 单元格渲染；缺省则按 accessor 取值 */
  render?: (row: T, rowIndex: number) => ReactNode;
  /** 取值器（用于排序 / 默认渲染） */
  accessor?: (row: T) => string | number | null | undefined;
  align?: TColumnAlign;
  sortable?: boolean;
  className?: string;
  headerClassName?: string;
};

type TDataTableProps<T> = {
  columns: TDataTableColumn<T>[];
  data: T[];
  /** 行唯一 key */
  rowKey: (row: T, index: number) => string | number;
  loading?: boolean;
  emptyText?: string;
  onRowClick?: (row: T) => void;
  className?: string;
};

type TSortState = { key: string; desc: boolean } | null;

const alignClass: Record<TColumnAlign, string> = {
  left: "text-left",
  center: "text-center",
  right: "text-right",
};

/**
 * 通用表格（见 FRONTEND.md §2 common/data-table）：
 * 列配置 + 客户端排序 + 加载 / 空态。分页由外层（如 usePagedQuery）控制。
 */
export function DataTable<T>({
  columns,
  data,
  rowKey,
  loading = false,
  emptyText = "暂无数据",
  onRowClick,
  className,
}: TDataTableProps<T>) {
  const [sort, setSort] = useState<TSortState>(null);

  const sortedData = useMemo(() => {
    if (!sort) return data;
    const column = columns.find((c) => c.key === sort.key);
    if (!column?.accessor) return data;
    const accessor = column.accessor;
    return [...data].sort((a, b) => {
      const av = accessor(a);
      const bv = accessor(b);
      if (av == null && bv == null) return 0;
      if (av == null) return 1;
      if (bv == null) return -1;
      if (av < bv) return sort.desc ? 1 : -1;
      if (av > bv) return sort.desc ? -1 : 1;
      return 0;
    });
  }, [data, sort, columns]);

  function toggleSort(column: TDataTableColumn<T>) {
    if (!column.sortable || !column.accessor) return;
    setSort((prev) => {
      if (!prev || prev.key !== column.key) return { key: column.key, desc: false };
      if (!prev.desc) return { key: column.key, desc: true };
      return null;
    });
  }

  if (!loading && sortedData.length === 0) {
    return <EmptyState title={emptyText} />;
  }

  return (
    <div className={cn("rounded-lg border", className)}>
      <Table>
        <TableHeader>
          <TableRow>
            {columns.map((column) => {
              const isSorted = sort?.key === column.key;
              return (
                <TableHead
                  key={column.key}
                  className={cn(
                    alignClass[column.align ?? "left"],
                    column.sortable && column.accessor ? "cursor-pointer select-none" : "",
                    column.headerClassName,
                  )}
                  onClick={() => toggleSort(column)}
                  aria-sort={isSorted ? (sort?.desc ? "descending" : "ascending") : undefined}
                >
                  <span className="inline-flex items-center gap-1">
                    {column.header}
                    {isSorted ? <span aria-hidden>{sort?.desc ? "▼" : "▲"}</span> : null}
                  </span>
                </TableHead>
              );
            })}
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading ? (
            <TableRow>
              <TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">
                加载中...
              </TableCell>
            </TableRow>
          ) : (
            sortedData.map((row, rowIndex) => (
              <TableRow
                key={rowKey(row, rowIndex)}
                className={onRowClick ? "cursor-pointer" : undefined}
                onClick={onRowClick ? () => onRowClick(row) : undefined}
              >
                {columns.map((column) => (
                  <TableCell
                    key={column.key}
                    className={cn(alignClass[column.align ?? "left"], column.className)}
                  >
                    {column.render
                      ? column.render(row, rowIndex)
                      : (column.accessor?.(row) ?? "--")}
                  </TableCell>
                ))}
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
