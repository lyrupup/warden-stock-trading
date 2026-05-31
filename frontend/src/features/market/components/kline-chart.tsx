import { useEffect, useRef } from "react";
import { createChart, type IChartApi } from "lightweight-charts";
import { useThemeStore } from "@/stores/theme-store";
import type { TKline } from "../types";

type TKlineChartProps = {
  data: TKline[];
  height?: number;
};

/** K 线图（lightweight-charts，见 FRONTEND.md §6 M1 个股详情） */
export function KlineChart({ data, height = 360 }: TKlineChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const theme = useThemeStore((s) => s.theme);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const isDark = theme === "dark";
    const chart: IChartApi = createChart(container, {
      height,
      width: container.clientWidth,
      layout: {
        background: { color: "transparent" },
        textColor: isDark ? "#cbd5e1" : "#334155",
      },
      grid: {
        vertLines: { color: isDark ? "#1e293b" : "#e2e8f0" },
        horzLines: { color: isDark ? "#1e293b" : "#e2e8f0" },
      },
      rightPriceScale: { borderColor: isDark ? "#334155" : "#cbd5e1" },
      timeScale: { borderColor: isDark ? "#334155" : "#cbd5e1" },
    });

    // A 股涨红跌绿
    const candleSeries = chart.addCandlestickSeries({
      upColor: "#dc2626",
      downColor: "#16a34a",
      borderUpColor: "#dc2626",
      borderDownColor: "#16a34a",
      wickUpColor: "#dc2626",
      wickDownColor: "#16a34a",
    });

    candleSeries.setData(
      data.map((k) => ({
        time: k.date,
        open: k.open,
        high: k.high,
        low: k.low,
        close: k.close,
      })),
    );
    chart.timeScale().fitContent();

    const handleResize = () => chart.applyOptions({ width: container.clientWidth });
    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
      chart.remove();
    };
  }, [data, height, theme]);

  return <div ref={containerRef} className="w-full" />;
}
