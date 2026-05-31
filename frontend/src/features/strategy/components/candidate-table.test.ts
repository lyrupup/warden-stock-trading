import { describe, expect, it } from "vitest";
import { factorMeta } from "./candidate-table";

describe("factorMeta", () => {
  it("固定均线/收盘走数值显示", () => {
    expect(factorMeta("ma5")).toEqual({ label: "MA5", kind: "number" });
    expect(factorMeta("close")).toEqual({ label: "收盘", kind: "number" });
  });

  it("乖离/振幅/涨跌幅走百分比显示", () => {
    expect(factorMeta("bias5")).toEqual({ label: "乖离5", kind: "percent" });
    expect(factorMeta("amplitude")).toEqual({ label: "振幅", kind: "percent" });
    expect(factorMeta("pct_change_5")).toEqual({ label: "涨跌幅", kind: "percent" });
  });

  it("ma_align 与 amplitude_streak 是布尔因子", () => {
    expect(factorMeta("ma_align_5_10_20").kind).toBe("bool");
    expect(factorMeta("ma_align_5_10_20").label).toBe("排列");
    expect(factorMeta("amplitude_streak").kind).toBe("bool");
  });

  it("vol_ratio 走倍数显示，不加单位", () => {
    expect(factorMeta("vol_ratio_5")).toEqual({ label: "量比", kind: "ratio" });
  });

  it("未知 key 退回原始名 + 数值显示", () => {
    expect(factorMeta("xyz")).toEqual({ label: "xyz", kind: "number" });
  });
});
