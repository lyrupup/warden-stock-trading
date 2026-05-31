import { describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { TPageResult } from "@/types";
import { createWrapper } from "@/test/test-utils";
import { usePagedQuery } from "./use-paged-query";

function makePage(page: number, size: number): TPageResult<{ id: number }> {
  const total = 45;
  const list = Array.from({ length: size }, (_, i) => ({ id: (page - 1) * size + i + 1 }));
  return { list, total, page, size };
}

describe("usePagedQuery", () => {
  it("初始加载第一页并计算 totalPages", async () => {
    const fetcher = vi.fn((page: number, size: number) => Promise.resolve(makePage(page, size)));
    const { result } = renderHook(() => usePagedQuery(["demo"], fetcher), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.page).toBe(1);
    expect(result.current.size).toBe(20);
    expect(result.current.total).toBe(45);
    expect(result.current.totalPages).toBe(3);
    expect(fetcher).toHaveBeenCalledWith(1, 20);
  });

  it("setPage 触发以新页码重新请求", async () => {
    const fetcher = vi.fn((page: number, size: number) => Promise.resolve(makePage(page, size)));
    const { result } = renderHook(() => usePagedQuery(["demo"], fetcher), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    act(() => result.current.setPage(2));

    await waitFor(() => expect(result.current.page).toBe(2));
    await waitFor(() => expect(fetcher).toHaveBeenCalledWith(2, 20));
  });
});
