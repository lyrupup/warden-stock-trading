import { useQuery } from "@tanstack/react-query";
import { marketApi } from "../api";

/**
 * 股票搜索：按代码/名称关键字查询，关键字非空时才发起请求。
 * 结果用于「行情中心·个股行情」tab 的搜索选择。
 */
export function useStockSearch(kw: string) {
  const keyword = kw.trim();
  return useQuery({
    queryKey: ["market", "search", keyword],
    queryFn: () => marketApi.search(keyword),
    enabled: keyword.length > 0,
  });
}
