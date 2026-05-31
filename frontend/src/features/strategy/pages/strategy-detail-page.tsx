import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/cn";
import { IndicatorBuilder } from "../components/indicator-builder";
import { ScreenPanel } from "../components/screen-panel";
import {
  useIndicatorCatalog,
  useSaveSkill,
  useStrategy,
  useUpdateIndicators,
  useUpdateStrategy,
} from "../hooks";
import type { TIndicatorCatalogItem, TIndicatorGroup } from "../types";

type TTab = "indicators" | "screen" | "skill" | "info";

const TABS: { key: TTab; label: string }[] = [
  { key: "indicators", label: "指标定义" },
  { key: "screen", label: "选股粗筛" },
  { key: "skill", label: "skill.md" },
  { key: "info", label: "基本信息" },
];

const emptyGroup: TIndicatorGroup = { logic: "and", rules: [] };

/** 策略详情 /strategies/:id（M2） */
export function StrategyDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const strategyId = Number(id);

  const [tab, setTab] = useState<TTab>("indicators");
  const detail = useStrategy(Number.isFinite(strategyId) ? strategyId : undefined);
  const catalog = useIndicatorCatalog();

  if (detail.isError) {
    return <EmptyState title="策略不存在或加载失败" actionLabel="返回列表" onAction={() => navigate("/strategies")} />;
  }

  const strategy = detail.data;
  const hasIndicators = !!strategy?.indicators && (strategy.indicators.rules?.length ?? 0) > 0;

  return (
    <div>
      <PageHeader
        title={strategy?.name ?? "策略详情"}
        description={strategy?.description || "量化指标定义 / 选股粗筛 / skill.md"}
        actions={
          <Button variant="outline" size="sm" onClick={() => navigate("/strategies")}>
            返回列表
          </Button>
        }
      />

      <div className="mb-6 flex gap-1 border-b">
        {TABS.map((tabItem) => (
          <button
            key={tabItem.key}
            type="button"
            className={cn(
              "border-b-2 px-4 py-2 text-sm transition-colors",
              tab === tabItem.key
                ? "border-primary font-medium text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
            onClick={() => setTab(tabItem.key)}
          >
            {tabItem.label}
          </button>
        ))}
      </div>

      {!strategy ? (
        <p className="text-sm text-muted-foreground">加载中…</p>
      ) : tab === "indicators" ? (
        <IndicatorTab strategyId={strategyId} initial={strategy.indicators ?? emptyGroup} catalogLoading={catalog.isLoading} catalog={catalog.data ?? []} />
      ) : tab === "screen" ? (
        <ScreenPanel strategyId={strategyId} hasIndicators={hasIndicators} />
      ) : tab === "skill" ? (
        <SkillTab strategyId={strategyId} initial={strategy.skill ?? ""} version={strategy.skill_version} />
      ) : (
        <InfoTab
          strategyId={strategyId}
          initial={{ name: strategy.name, description: strategy.description, tags: strategy.tags }}
        />
      )}
    </div>
  );
}

type TIndicatorTabProps = {
  strategyId: number;
  initial: TIndicatorGroup;
  catalog: TIndicatorCatalogItem[];
  catalogLoading: boolean;
};

function IndicatorTab({ strategyId, initial, catalog, catalogLoading }: TIndicatorTabProps) {
  const [group, setGroup] = useState<TIndicatorGroup>(initial);
  const save = useUpdateIndicators(strategyId);

  useEffect(() => {
    setGroup(initial);
    // 仅在策略切换/首次加载时同步初值。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [strategyId]);

  if (catalogLoading) return <p className="text-sm text-muted-foreground">因子目录加载中…</p>;

  return (
    <div className="space-y-4">
      <IndicatorBuilder value={group} catalog={catalog} onChange={setGroup} />
      <div className="flex items-center gap-3">
        <Button onClick={() => save.mutate(group)} disabled={save.isPending || group.rules.length === 0}>
          保存指标定义
        </Button>
        {save.isSuccess ? <span className="text-sm text-quote-down">已保存</span> : null}
        {save.isError ? (
          <span className="text-sm text-quote-down">{(save.error as Error)?.message ?? "保存失败"}</span>
        ) : null}
      </div>
    </div>
  );
}

function SkillTab({ strategyId, initial, version }: { strategyId: number; initial: string; version?: number }) {
  const [content, setContent] = useState(initial);
  const save = useSaveSkill(strategyId);

  useEffect(() => {
    setContent(initial);
  }, [initial]);

  return (
    <div className="space-y-3">
      <p className="text-xs text-muted-foreground">
        策略 skill.md（驱动 M6 AI 精筛的提示词框架）{version ? `· 当前版本 v${version}` : ""}
      </p>
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        className="min-h-[24rem] w-full rounded-md border border-input bg-background p-3 font-mono text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring"
        placeholder="# 分析框架&#10;请基于本策略对个股进行..."
      />
      <Button onClick={() => save.mutate(content)} disabled={save.isPending}>
        保存（生成新版本）
      </Button>
    </div>
  );
}

function InfoTab({
  strategyId,
  initial,
}: {
  strategyId: number;
  initial: { name: string; description: string; tags: string };
}) {
  const [form, setForm] = useState(initial);
  const save = useUpdateStrategy(strategyId);

  useEffect(() => {
    setForm(initial);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [strategyId]);

  return (
    <div className="max-w-xl space-y-4">
      <label className="flex flex-col gap-1 text-sm">
        <span className="text-muted-foreground">名称</span>
        <Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <span className="text-muted-foreground">描述</span>
        <textarea
          value={form.description}
          onChange={(e) => setForm({ ...form, description: e.target.value })}
          className="min-h-[8rem] w-full rounded-md border border-input bg-background p-3 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring"
        />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <span className="text-muted-foreground">标签（逗号分隔）</span>
        <Input value={form.tags} onChange={(e) => setForm({ ...form, tags: e.target.value })} />
      </label>
      <Button onClick={() => save.mutate(form)} disabled={save.isPending || !form.name.trim()}>
        保存
      </Button>
    </div>
  );
}
