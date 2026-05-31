import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useDebounce } from "@/hooks/use-debounce";
import {
  useCopyStrategy,
  useCreateStrategy,
  useDeleteStrategy,
  useStrategies,
  useStrategyTemplates,
} from "../hooks";
import type { TStrategyTemplate } from "../types";

/** 策略列表 /strategies（M2） */
export function StrategyListPage() {
  const navigate = useNavigate();
  const [kw, setKw] = useState("");
  const debouncedKw = useDebounce(kw, 300);

  const listQuery = useStrategies(debouncedKw ? { kw: debouncedKw } : undefined);
  const templatesQuery = useStrategyTemplates();
  const createStrategy = useCreateStrategy();
  const copyStrategy = useCopyStrategy();
  const deleteStrategy = useDeleteStrategy();

  function handleCreateBlank() {
    createStrategy.mutate(
      { name: "新建策略" },
      { onSuccess: (s) => navigate(`/strategies/${s.id}`) },
    );
  }

  function handleCreateFromTemplate(tpl: TStrategyTemplate) {
    createStrategy.mutate(
      { name: tpl.name, description: tpl.description, tags: tpl.tags, indicators: tpl.indicators },
      { onSuccess: (s) => navigate(`/strategies/${s.id}`) },
    );
  }

  const strategies = listQuery.data ?? [];

  return (
    <div>
      <PageHeader
        title="选股策略"
        description="定义量化指标策略，对股票池做量化粗筛，产出候选池"
        actions={
          <Button onClick={handleCreateBlank} disabled={createStrategy.isPending}>
            新建策略
          </Button>
        }
      />

      <section className="mb-8">
        <h2 className="mb-3 text-sm font-medium text-muted-foreground">从模板快速创建</h2>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {(templatesQuery.data ?? []).map((tpl) => (
            <Card key={tpl.key}>
              <CardHeader>
                <CardTitle className="text-base">{tpl.name}</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <p className="text-sm text-muted-foreground">{tpl.description}</p>
                <p className="text-xs text-muted-foreground">适用：{tpl.scenario}</p>
              </CardContent>
              <CardFooter>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleCreateFromTemplate(tpl)}
                  disabled={createStrategy.isPending}
                >
                  使用此模板
                </Button>
              </CardFooter>
            </Card>
          ))}
        </div>
      </section>

      <section>
        <div className="mb-3 flex items-center justify-between gap-4">
          <h2 className="text-sm font-medium text-muted-foreground">我的策略</h2>
          <Input
            value={kw}
            onChange={(e) => setKw(e.target.value)}
            placeholder="搜索策略名称"
            className="h-8 w-48"
          />
        </div>

        {listQuery.isError ? (
          <EmptyState title="加载失败" actionLabel="重试" onAction={() => void listQuery.refetch()} />
        ) : strategies.length === 0 ? (
          <EmptyState title="还没有策略，点击右上角「新建策略」或使用模板创建" />
        ) : (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {strategies.map((s) => (
              <Card
                key={s.id}
                className="cursor-pointer transition-colors hover:border-primary/50"
                onClick={() => navigate(`/strategies/${s.id}`)}
              >
                <CardHeader>
                  <CardTitle className="text-base">{s.name}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <p className="line-clamp-2 text-sm text-muted-foreground">{s.description || "暂无描述"}</p>
                  <div className="flex flex-wrap gap-1">
                    {s.tags
                      ? s.tags
                          .split(",")
                          .filter(Boolean)
                          .map((tag) => (
                            <Badge key={tag} variant="secondary">
                              {tag}
                            </Badge>
                          ))
                      : null}
                  </div>
                </CardContent>
                <CardFooter className="gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      copyStrategy.mutate(s.id);
                    }}
                  >
                    复制
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      deleteStrategy.mutate(s.id);
                    }}
                  >
                    删除
                  </Button>
                </CardFooter>
              </Card>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
