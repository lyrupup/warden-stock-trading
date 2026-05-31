import { useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import type { TPageResult } from "@/types";

/**
 * 通用分页 Hook（见 FRONTEND.md §4.3）：抽象「分页 + 加载 + 错误」逻辑，
 * 供持仓流水、报告列表、任务日志等列表页复用。
 */
export function usePagedQuery<T>(
  key: unknown[],
  fetcher: (page: number, size: number) => Promise<TPageResult<T>>,
  initialSize = 20,
) {
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(initialSize);

  const query = useQuery({
    queryKey: [...key, page, size],
    queryFn: () => fetcher(page, size),
    placeholderData: keepPreviousData,
  });

  const total = query.data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / size));

  return {
    ...query,
    page,
    size,
    total,
    totalPages,
    setPage,
    setSize,
  };
}
