import { create } from "zustand";

export type TLocale = "zh-CN" | "en-US";

/**
 * 用户配置缓存（M7 个人配置中的非敏感部分），供全局读取。
 * 敏感字段（API Key 等）不在前端缓存，仅由后端脱敏返回。
 */
export type TUserConfigState = {
  /** 行情刷新频率（毫秒），默认 5s */
  marketRefreshMs: number;
  /** 语言偏好 */
  locale: TLocale;
  /** 首页默认视图路由 */
  defaultView: string;
  setMarketRefreshMs: (ms: number) => void;
  setLocale: (locale: TLocale) => void;
  setDefaultView: (path: string) => void;
};

export const useUserConfigStore = create<TUserConfigState>((set) => ({
  marketRefreshMs: 5000,
  locale: "zh-CN",
  defaultView: "/dashboard",
  setMarketRefreshMs: (ms) => set({ marketRefreshMs: ms }),
  setLocale: (locale) => set({ locale }),
  setDefaultView: (path) => set({ defaultView: path }),
}));
