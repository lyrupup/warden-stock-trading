import { http, HttpResponse } from "msw";
import {
  addMockWatch,
  genMockKline,
  getMockWatchlistQuotes,
  mockIndices,
  mockQuotes,
  mockWatchlist,
  removeMockWatch,
} from "./data";

/** 统一成功响应体 {code,message,data} */
function ok<T>(data: T) {
  return HttpResponse.json({ code: 0, message: "ok", data });
}

/**
 * MSW handlers：模拟后端 /market 等接口（见 FRONTEND.md §8）。
 * 用通配前缀（星号 + 路径）匹配，兼容不同 VITE_API_BASE_URL。
 */
export const handlers = [
  // ---------------- Auth ----------------
  http.post("*/auth/login", () =>
    ok({ token: "dev-single-user-token", user: { id: 1, username: "warden", nickname: "守望者" } }),
  ),
  http.get("*/auth/me", () => ok({ id: 1, username: "warden", nickname: "守望者", status: 1 })),

  // ---------------- Market ----------------
  http.get("*/market/indices", () => ok(mockIndices)),
  http.get("*/market/overview", () =>
    ok({ up_count: 2456, down_count: 1789, limit_up: 38, limit_down: 12, northbound: 2.34e9 }),
  ),
  http.get("*/market/watchlist", () => ok(mockWatchlist)),
  http.get("*/market/watchlist/quotes", () => ok(getMockWatchlistQuotes())),
  http.post("*/market/watchlist", async ({ request }) => {
    const body = (await request.json()) as { stock_code: string; group_name?: string; remark?: string };
    const item = addMockWatch(body.stock_code, body.group_name, body.remark);
    return ok(item);
  }),
  http.delete("*/market/watchlist/:id", ({ params }) => {
    removeMockWatch(Number(params.id));
    return ok(null);
  }),
  http.get("*/market/stocks/:code/kline", ({ params }) => {
    const code = String(params.code);
    const seed = mockQuotes[code]?.price ?? 100;
    return ok(genMockKline(seed));
  }),
  http.get("*/market/stocks/:code", ({ params }) => {
    const code = String(params.code);
    const quote = mockQuotes[code];
    if (!quote) return HttpResponse.json({ code: 40400, message: "stock not found", data: null });
    return ok(quote);
  }),
  http.get("*/market/search", ({ request }) => {
    const kw = new URL(request.url).searchParams.get("kw") ?? "";
    const list = Object.values(mockQuotes).filter(
      (q) => q.stock_code.includes(kw) || q.stock_name.includes(kw),
    );
    return ok(list);
  }),
];
