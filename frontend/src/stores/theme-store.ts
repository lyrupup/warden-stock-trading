import { create } from "zustand";

export type TTheme = "light" | "dark";

const THEME_KEY = "warden_theme";

function applyThemeClass(theme: TTheme): void {
  if (typeof document === "undefined") return;
  const root = document.documentElement;
  root.classList.toggle("dark", theme === "dark");
}

function getInitialTheme(): TTheme {
  const saved = (typeof localStorage !== "undefined"
    ? localStorage.getItem(THEME_KEY)
    : null) as TTheme | null;
  if (saved === "light" || saved === "dark") return saved;
  if (typeof window !== "undefined" && window.matchMedia?.("(prefers-color-scheme: dark)").matches) {
    return "dark";
  }
  return "light";
}

/** 主题状态：light/dark 切换 + 持久化 + 切换 html class（见 FRONTEND.md §4.4/§4.5） */
export type TThemeState = {
  theme: TTheme;
  setTheme: (theme: TTheme) => void;
  toggleTheme: () => void;
};

export const useThemeStore = create<TThemeState>((set, get) => ({
  theme: getInitialTheme(),
  setTheme: (theme) => {
    localStorage.setItem(THEME_KEY, theme);
    applyThemeClass(theme);
    set({ theme });
  },
  toggleTheme: () => {
    const next: TTheme = get().theme === "dark" ? "light" : "dark";
    get().setTheme(next);
  },
}));

/** 应用启动时同步一次 html class（在 main.tsx 调用） */
export function initThemeClass(): void {
  applyThemeClass(useThemeStore.getState().theme);
}
