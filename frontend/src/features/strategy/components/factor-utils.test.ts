import { describe, expect, it } from "vitest";
import type { TIndicatorCatalogItem } from "../types";
import { defaultFactor, describeFactor, isBoolFactor } from "./factor-utils";

const catalog: TIndicatorCatalogItem[] = [
  {
    type: "ma",
    name: "移动均线 MA",
    value_type: "number",
    description: "",
    params: [{ key: "period", type: "int", required: true, default: 5, desc: "" }],
  },
  {
    type: "ma_align",
    name: "均线排列",
    value_type: "bool",
    description: "",
    params: [
      { key: "periods", type: "int[]", required: true, default: [5, 10, 20], desc: "" },
      { key: "direction", type: "enum", required: true, default: "bull", enum: ["bull", "bear"], desc: "" },
      { key: "eps", type: "float", required: false, default: 0, desc: "" },
    ],
  },
];

describe("defaultFactor", () => {
  it("依据目录默认参数构造 ma 因子", () => {
    expect(defaultFactor("ma", catalog)).toEqual({ type: "ma", period: 5 });
  });

  it("构造 ma_align 因子带有序周期与方向", () => {
    const f = defaultFactor("ma_align", catalog);
    expect(f.type).toBe("ma_align");
    expect(f.periods).toEqual([5, 10, 20]);
    expect(f.direction).toBe("bull");
    expect(f.eps).toBe(0);
  });

  it("const 因子默认值为 0", () => {
    expect(defaultFactor("const", catalog)).toEqual({ type: "const", value: 0 });
  });
});

describe("describeFactor", () => {
  it("多头排列描述用 > 连接", () => {
    expect(describeFactor({ type: "ma_align", periods: [5, 10, 20], direction: "bull" })).toContain("5>10>20");
  });

  it("空头排列描述用 < 连接", () => {
    expect(describeFactor({ type: "ma_align", periods: [20, 30, 60], direction: "bear" })).toContain("20<30<60");
  });

  it("乖离率与振幅描述", () => {
    expect(describeFactor({ type: "bias", period: 5 })).toBe("乖离率(5)");
    expect(describeFactor({ type: "amplitude" })).toBe("振幅");
    expect(describeFactor({ type: "amplitude_streak", threshold: 5, days: 3 })).toContain("连续振幅≥5%×3");
  });
});

describe("isBoolFactor", () => {
  it("ma_align / amplitude_streak 为布尔因子", () => {
    expect(isBoolFactor("ma_align")).toBe(true);
    expect(isBoolFactor("amplitude_streak")).toBe(true);
    expect(isBoolFactor("ma")).toBe(false);
  });
});
