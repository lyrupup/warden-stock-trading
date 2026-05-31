import { useEffect, useMemo, useRef, useState } from "react";
import {
  createChart,
  type BarData,
  type IChartApi,
  type ISeriesApi,
  type LineData,
  type LineWidth,
  type MouseEventParams,
  type Time,
} from "lightweight-charts";
import { useTranslation } from "react-i18next";
import { useThemeStore } from "@/stores/theme-store";
import { cn } from "@/lib/cn";
import { formatPrice, getQuoteColor } from "@/lib/format";
import type { TKline } from "../types";

/** 可叠加的 MA 周期，覆盖国内常用 5/10/20/30/60/120 日均线 */
export const MA_PERIODS = [5, 10, 20, 30, 60, 120] as const;
export type TMAPeriod = (typeof MA_PERIODS)[number];

/**
 * MA 系列配色，对齐同花顺/通达信经典：白→黄→紫→绿→蓝→红。
 * 调整点：
 * - MA5  浅色主题下"白"不可见，改用同样高对比的醒目橙
 * - MA30 深绿（green-700）避开 K 线跌绿 #16a34a 撞色
 * - MA120 同花顺多为红/品红，避开 K 线涨红 #dc2626 后用更偏品的玫红
 * 与开关 chip 上的圆点颜色一致。
 */
export const MA_COLOR: Record<TMAPeriod, string> = {
  5: "#ea580c",
  10: "#eab308",
  20: "#a855f7",
  30: "#15803d",
  60: "#2563eb",
  120: "#be185d",
};

/** A 股配色：涨红跌绿 */
const UP_COLOR = "#dc2626";
const DOWN_COLOR = "#16a34a";

/**
 * lightweight-charts 的 LineWidth 类型限定为 1|2|3|4，
 * 但底层 canvas strokeWidth 接受小数；这里用 1.5 以兼顾纤细与可读。
 */
const MA_LINE_WIDTH = 1.5 as unknown as LineWidth;

/** 收盘价简单移动平均；样本不足该期不输出，避免首端误导性均值 */
export function computeMA(data: TKline[], period: number): { time: string; value: number }[] {
  if (period <= 0 || data.length < period) return [];
  const out: { time: string; value: number }[] = [];
  let sum = 0;
  for (let i = 0; i < data.length; i++) {
    sum += data[i].close;
    if (i >= period) sum -= data[i - period].close;
    if (i >= period - 1) out.push({ time: data[i].date, value: +(sum / period).toFixed(4) });
  }
  return out;
}

type TKlineChartProps = {
  data: TKline[];
  height?: number;
  /** 叠加 MA */
  enabledMAs?: readonly TMAPeriod[];
  /** 初始可视 K 线根数；超过总数则按总数 fit。默认 60（约 3 个月日线） */
  initialVisibleBars?: number;
};

/** K 线图（lightweight-charts，可叠加 MA + hover OHLC 浮层，见 FRONTEND.md §6 M1） */
export function KlineChart({
  data,
  height = 360,
  enabledMAs = [],
  initialVisibleBars = 60,
}: TKlineChartProps) {
  const { t } = useTranslation();
  const containerRef = useRef<HTMLDivElement>(null);
  const theme = useThemeStore((s) => s.theme);

  // hover 行情；null 时使用最末根 K 线，使浮层始终可见
  const [hover, setHover] = useState<TKline | null>(null);
  const latest = data.length > 0 ? data[data.length - 1] : null;
  const display = hover ?? latest;

  // 数组引用每次都新——用排序后的 key 让 effect 稳定地依赖 MA 集合
  const maKey = useMemo(
    () => [...enabledMAs].sort((a, b) => a - b).join(","),
    [enabledMAs],
  );

  // 浮层用的 MA 取值表：按日期反查每个已勾选 MA 的当日值；样本不足则为 undefined
  const maMaps = useMemo(() => {
    const out: Partial<Record<TMAPeriod, Map<string, number>>> = {};
    for (const p of enabledMAs) {
      out[p] = new Map(computeMA(data, p).map((it) => [it.time, it.value]));
    }
    return out;
    // 仅依赖数据本体与 enabled 集合的内容；引用变化由 maKey 捕获
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data, maKey]);

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

    const candleSeries = chart.addCandlestickSeries({
      upColor: UP_COLOR,
      downColor: DOWN_COLOR,
      borderUpColor: UP_COLOR,
      borderDownColor: DOWN_COLOR,
      wickUpColor: UP_COLOR,
      wickDownColor: DOWN_COLOR,
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

    // 叠加 MA 折线
    const periods = [...enabledMAs].sort((a, b) => a - b);
    for (const p of periods) {
      const s = chart.addLineSeries({
        color: MA_COLOR[p],
        lineWidth: MA_LINE_WIDTH,
        priceLineVisible: false,
        lastValueVisible: false,
        crosshairMarkerVisible: false,
      });
      s.setData(computeMA(data, p) as LineData<Time>[]);
    }

    // 初始可视窗口：只展示最末 N 根，避免一打开 K 线被压扁
    const bars = Math.min(initialVisibleBars, data.length);
    if (bars > 0 && bars < data.length) {
      chart.timeScale().setVisibleLogicalRange({
        from: data.length - bars - 0.5,
        to: data.length - 0.5,
      });
    } else {
      chart.timeScale().fitContent();
    }

    // 用 date 作为索引，便于 crosshair 中按 time 反查完整 K 线
    const byDate = new Map(data.map((k) => [k.date, k]));
    const handler = (param: MouseEventParams<Time>) => {
      if (!param.time) {
        setHover(null);
        return;
      }
      // candleSeries 在 seriesData 中存的是 BarData，含 time/open/high/low/close
      const bar = param.seriesData.get(
        candleSeries as ISeriesApi<"Candlestick", Time>,
      ) as BarData<Time> | undefined;
      if (bar) {
        const k = byDate.get(String(bar.time));
        if (k) {
          setHover(k);
          return;
        }
        // 兜底：用 seriesData 自身重组（一般不会走到）
        setHover({
          date: String(bar.time),
          open: bar.open,
          high: bar.high,
          low: bar.low,
          close: bar.close,
          volume: 0,
          amount: 0,
        });
      } else {
        setHover(null);
      }
    };
    chart.subscribeCrosshairMove(handler);

    const handleResize = () => chart.applyOptions({ width: container.clientWidth });
    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
      chart.unsubscribeCrosshairMove(handler);
      chart.remove();
    };
    // maKey 已覆盖 enabledMAs 内容变化
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data, height, theme, maKey, initialVisibleBars]);

  // 浮层按 enabled 顺序输出（再按周期升序保证视觉稳定）
  const maItems = useMemo(
    () =>
      [...enabledMAs]
        .sort((a, b) => a - b)
        .map((p) => ({
          period: p,
          value: display ? maMaps[p]?.get(display.date) : undefined,
        })),
    [enabledMAs, display, maMaps],
  );

  return (
    <div>
      {display ? (
        <div className="mb-2 space-y-1">
          <OhlcStrip
            k={display}
            isHover={hover !== null}
            labels={{
              open: t("market.columns.open"),
              close: t("market.columns.close"),
              high: t("market.columns.high"),
              low: t("market.columns.low"),
            }}
          />
          {maItems.length > 0 ? <MaStrip items={maItems} /> : null}
        </div>
      ) : null}
      <div ref={containerRef} className="w-full" onMouseLeave={() => setHover(null)} />
    </div>
  );
}

/** K 线上方 OHLC 浮层：默认显示最末根；hover 时切换到光标所在 K 线 */
function OhlcStrip({
  k,
  isHover,
  labels,
}: {
  k: TKline;
  isHover: boolean;
  labels: { open: string; close: string; high: string; low: string };
}) {
  const change = k.close - k.open;
  const color = getQuoteColor(change);
  const pct = k.open === 0 ? 0 : (change / k.open) * 100;
  const sign = change >= 0 ? "+" : "";

  return (
    <div className="mb-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs tabular-nums">
      <span className={cn("text-muted-foreground", isHover ? "font-medium text-foreground" : "")}>
        {k.date}
      </span>
      <Item label={labels.open} value={k.open} color={color} />
      <Item label={labels.high} value={k.high} color={color} />
      <Item label={labels.low} value={k.low} color={color} />
      <Item label={labels.close} value={k.close} color={color} />
      <span className={color}>
        {sign}
        {formatPrice(change)} ({sign}
        {pct.toFixed(2)}%)
      </span>
    </div>
  );
}

function Item({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <span>
      <span className="text-muted-foreground">{label} </span>
      <span className={cn("font-medium", color)}>{formatPrice(value)}</span>
    </span>
  );
}

/** K 线上方 MA 浮层：与 chart 上的 MA 线一一对应（同色 + 同 period） */
function MaStrip({ items }: { items: { period: TMAPeriod; value: number | undefined }[] }) {
  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs tabular-nums">
      {items.map(({ period, value }) => (
        <span key={period} className="font-medium" style={{ color: MA_COLOR[period] }}>
          MA{period} {value !== undefined ? formatPrice(value) : "--"}
        </span>
      ))}
    </div>
  );
}
