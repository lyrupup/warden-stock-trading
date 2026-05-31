/** 导航分区（按守望者四大技能彩蛋 + 设置，见 FRONTEND.md §5.2 / PRD §1.2） */
export type TNavItem = {
  to: string;
  labelKey: string;
};

export type TNavSection = {
  titleKey: string;
  items: TNavItem[];
};

export const NAV_SECTIONS: TNavSection[] = [
  {
    titleKey: "nav.section.scan",
    items: [
      { to: "/market", labelKey: "nav.market" },
      { to: "/positions", labelKey: "nav.positions" },
    ],
  },
  {
    titleKey: "nav.section.insight",
    items: [
      { to: "/strategies", labelKey: "nav.strategies" },
      { to: "/ai", labelKey: "nav.ai" },
      { to: "/ai/reports", labelKey: "nav.reports" },
    ],
  },
  {
    titleKey: "nav.section.alert",
    items: [
      { to: "/risk", labelKey: "nav.risk" },
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
