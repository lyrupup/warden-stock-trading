import { NavLink, Outlet } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { cn } from "@/lib/cn";
import { Button } from "@/components/ui/button";
import { useThemeStore } from "@/stores/theme-store";
import { NAV_SECTIONS } from "./nav-config";

/** 全局布局：左侧分区导航 + 顶栏（主题切换） + 内容区（见 FRONTEND.md §5.2） */
export function AppLayout() {
  const { t } = useTranslation();
  const theme = useThemeStore((s) => s.theme);
  const toggleTheme = useThemeStore((s) => s.toggleTheme);

  return (
    <div className="flex min-h-screen bg-background text-foreground">
      <aside className="hidden w-60 shrink-0 border-r md:flex md:flex-col">
        <div className="flex h-14 items-center gap-2 border-b px-4">
          <span className="text-lg">🛡️</span>
          <span className="font-semibold">{t("app.name")}</span>
        </div>
        <nav className="flex-1 space-y-4 overflow-y-auto p-3">
          {NAV_SECTIONS.map((section) => (
            <div key={section.titleKey}>
              <div className="px-2 pb-1 text-xs font-medium uppercase tracking-wider text-muted-foreground">
                {t(section.titleKey)}
              </div>
              <div className="space-y-0.5">
                {section.items.map((item) => (
                  <NavLink
                    key={item.to}
                    to={item.to}
                    className={({ isActive }) =>
                      cn(
                        "block rounded-md px-3 py-2 text-sm transition-colors",
                        isActive
                          ? "bg-accent font-medium text-accent-foreground"
                          : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
                      )
                    }
                  >
                    {t(item.labelKey)}
                  </NavLink>
                ))}
              </div>
            </div>
          ))}
        </nav>
      </aside>

      <div className="flex flex-1 flex-col">
        <header className="flex h-14 items-center justify-between border-b px-4">
          <NavLink to="/dashboard" className="text-sm font-medium">
            {t("nav.dashboard")}
          </NavLink>
          <Button
            variant="ghost"
            size="icon"
            onClick={toggleTheme}
            aria-label={t("theme.toggle")}
            title={t("theme.toggle")}
          >
            {theme === "dark" ? "🌙" : "☀️"}
          </Button>
        </header>
        <main className="flex-1 overflow-y-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
