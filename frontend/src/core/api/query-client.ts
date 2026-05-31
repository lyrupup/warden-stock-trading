import { QueryClient } from "@tanstack/react-query";
import { AppError } from "@/core/http-client";

/** 全局 TanStack Query 客户端 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: (failureCount, error) => {
        // 业务错误（AppError）不重试，网络错误最多重试 1 次
        if (error instanceof AppError) return false;
        return failureCount < 1;
      },
      refetchOnWindowFocus: false,
    },
  },
});
