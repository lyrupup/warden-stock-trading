import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { PageHeader } from "@/components/common/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const ENTRIES = [
  { to: "/market", titleKey: "nav.market", section: "🔍 侦查" },
  { to: "/strategies", titleKey: "nav.strategies", section: "🧠 洞察" },
  { to: "/positions", titleKey: "nav.positions", section: "🔍 侦查" },
  { to: "/risk", titleKey: "nav.risk", section: "🚨 警戒" },
  { to: "/ai", titleKey: "nav.ai", section: "🧠 洞察" },
  { to: "/tasks", titleKey: "nav.tasks", section: "📨 传信" },
];

/** 工作台 /dashboard（聚合入口） */
export function DashboardPage() {
  const { t } = useTranslation();
  return (
    <div>
      <PageHeader title={t("nav.dashboard")} description={t("app.disclaimer")} />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {ENTRIES.map((e) => (
          <Link key={e.to} to={e.to}>
            <Card className="transition-colors hover:bg-accent">
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <span className="text-base">{t(e.titleKey)}</span>
                  <span className="text-xs text-muted-foreground">{e.section}</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">进入 {t(e.titleKey)}</p>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
