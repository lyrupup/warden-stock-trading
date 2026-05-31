import { useEffect, type ReactNode } from "react";
import { useThemeStore } from "@/stores/theme-store";

type TThemeProviderProps = {
  children: ReactNode;
};

/**
 * 主题 Provider：订阅 theme-store，主题变化时同步 <html class="dark">。
 * Tailwind darkMode: "class"（见 FRONTEND.md §4.5）。
 */
export function ThemeProvider({ children }: TThemeProviderProps) {
  const theme = useThemeStore((s) => s.theme);

  useEffect(() => {
    const root = document.documentElement;
    root.classList.toggle("dark", theme === "dark");
  }, [theme]);

  return <>{children}</>;
}
