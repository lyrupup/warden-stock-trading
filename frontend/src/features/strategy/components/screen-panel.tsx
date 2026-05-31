import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAddWatch } from "@/features/market";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { useRunScreen } from "../hooks";
import type { TScreenResult, TScreenUniverse } from "../types";
import { CandidateTable } from "./candidate-table";

type TScreenPanelProps = {
  strategyId: number;
  /** 策略是否已定义指标，未定义时禁用并提示。 */
  hasIndicators: boolean;
};

const selectClass =
  "h-9 rounded-md border border-input bg-background px-2 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring";

const STATUS_TEXT: Record<number, string> = {
  0: "排队中…",
  1: "粗筛运行中…",
  2: "完成",
  3: "失败",
};

/** 选股粗筛面板：股票池选择 + 运行 + 候选结果表。 */
export function ScreenPanel({ strategyId, hasIndicators }: TScreenPanelProps) {
  const navigate = useNavigate();
  const addWatch = useAddWatch();
  const { run, result } = useRunScreen(strategyId);

  const [universeType, setUniverseType] = useState<TScreenUniverse["type"]>("watchlist");
  const [codesText, setCodesText] = useState("");
  const [limit, setLimit] = useState(200);

  const data: TScreenResult | undefined = result.data;
  const running = run.isPending || data?.status === 0 || data?.status === 1;

  function handleRun() {
    const universe: TScreenUniverse = { type: universeType };
    if (universeType === "codes") {
      universe.codes = codesText
        .split(/[\s,，]+/)
        .map((s) => s.trim())
        .filter(Boolean);
    }
    run.mutate({ universe, limit });
  }

  const unsupported = universeType === "all" || universeType === "board";

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="flex flex-wrap items-end gap-4 pt-6">
          <label className="flex flex-col gap-1 text-sm">
            <span className="text-muted-foreground">股票池</span>
            <select
              className={selectClass}
              value={universeType}
              onChange={(e) => setUniverseType(e.target.value as TScreenUniverse["type"])}
            >
              <option value="watchlist">自选股</option>
              <option value="codes">自定义代码</option>
              <option value="all">全市场（暂未支持）</option>
              <option value="board">板块（暂未支持）</option>
            </select>
          </label>

          {universeType === "codes" ? (
            <label className="flex min-w-[16rem] flex-1 flex-col gap-1 text-sm">
              <span className="text-muted-foreground">股票代码（逗号/空格分隔）</span>
              <Input
                value={codesText}
                onChange={(e) => setCodesText(e.target.value)}
                placeholder="600519, 000001, 300750"
                className="h-9"
              />
            </label>
          ) : null}

          <label className="flex w-28 flex-col gap-1 text-sm">
            <span className="text-muted-foreground">候选上限</span>
            <Input
              type="number"
              value={limit}
              onChange={(e) => setLimit(Number(e.target.value) || 200)}
              className="h-9"
            />
          </label>

          <Button onClick={handleRun} disabled={!hasIndicators || running || unsupported}>
            {running ? "粗筛中…" : "运行粗筛"}
          </Button>
        </CardContent>
      </Card>

      {!hasIndicators ? (
        <p className="rounded-md border border-dashed px-3 py-3 text-sm text-muted-foreground">
          该策略尚未定义量化指标，请先在「指标定义」中配置后再运行粗筛。
        </p>
      ) : null}

      {unsupported ? (
        <p className="rounded-md border border-dashed px-3 py-3 text-sm text-muted-foreground">
          全市场 / 板块粗筛需个股指标快照（后续迭代），当前请使用「自选股」或「自定义代码」。
        </p>
      ) : null}

      {run.isError ? (
        <p className="rounded-md border border-quote-down/40 bg-muted px-3 py-3 text-sm text-quote-down">
          粗筛失败：{(run.error as Error)?.message ?? "未知错误"}
        </p>
      ) : null}

      {data ? (
        <div className="space-y-2">
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <span>状态：{STATUS_TEXT[data.status] ?? "--"}</span>
            <span>扫描 {data.universe_count} 只</span>
            <span>命中 {data.matched_count} 只</span>
            {data.trade_date ? <span>基准日 {data.trade_date}</span> : null}
          </div>
          {data.status === 3 ? (
            <p className="text-sm text-quote-down">{data.error_msg}</p>
          ) : (
            <CandidateTable
              candidates={data.candidates}
              loading={result.isLoading}
              onAddWatch={(code) => addWatch.mutate({ stock_code: code })}
              onAnalyze={(code) => navigate(`/ai?strategy=${strategyId}&code=${code}`)}
            />
          )}
        </div>
      ) : null}
    </div>
  );
}
