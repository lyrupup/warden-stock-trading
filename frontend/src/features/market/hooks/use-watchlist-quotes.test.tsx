import { describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { createWrapper } from "@/test/test-utils";
import { useWatchlistQuotes } from "./use-watchlist-quotes";

describe("useWatchlistQuotes (MSW)", () => {
  it("通过 MSW mock 拉取自选股行情并解包 data", async () => {
    const { result } = renderHook(() => useWatchlistQuotes(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const quotes = result.current.data ?? [];
    expect(quotes.length).toBeGreaterThan(0);
    expect(quotes[0]).toHaveProperty("stock_code");
    expect(quotes.map((q) => q.stock_code)).toContain("600519");
  });
});
