/** 导航分区（按守望者四大技能彩蛋 + 设置，见 FRONTEND.md §5.2 / PRD §1.2） */
export type TNavItem = {
  to: string;
  labelKey: string;
  /** 精确匹配高亮：当本项路径是其它导航项的前缀时需开启，避免子路由把父项一起点亮 */
  end?: boolean;
};

export type TNavSection = {
  titleKey: string;
  items: TNavItem[];
};

export const NAV_SECTIONS: TNavSection[] = [
  {
    titleKey: "nav.section.scan",
    items: [
      { to: "/market", labelKey: "nav.market", end: true },
      { to: "/market/quote", labelKey: "nav.stockQuote" },
      { to: "/positions", labelKey: "nav.positions" },
    ],
  },
  {
    titleKey: "nav.section.insight",
    items: [
      { to: "/strategies", labelKey: "nav.strategies" },
      { to: "/ai", labelKey: "nav.ai", end: true },
      { to: "/ai/reports", labelKey: "nav.reports" },
    ],
  },
  {
    titleKey: "nav.section.alert",
    items: [
      { to: "/risk", labelKey: "nav.risk", end: true },
      { to: "/risk/premarket", labelKey: "nav.premarket" },
    ],
  },
  {
    titleKey: "nav.section.message",
    items: [{ to: "/tasks", labelKey: "nav.tasks" }],
  },
  {
    titleKey: "nav.section.system",
    items: [{ to: "/settings", labelKey: "nav.settings" }],
  },
];
