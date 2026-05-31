import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";

type TPlaceholderPageProps = {
  title: string;
  description?: string;
  /** 所属里程碑标记，如 "M2" */
  milestone?: string;
};

/**
 * 占位页：用于尚未实现的路由，保证骨架可跳转。
 * M1 行情已作为竖切范式完整实现，其余模块按此范式逐步替换。
 */
export function PlaceholderPage({ title, description, milestone }: TPlaceholderPageProps) {
  return (
    <div>
      <PageHeader title={title} description={description} />
      <EmptyState
        title={milestone ? `${milestone} · 待实现` : "待实现"}
        description="该模块为脚手架占位页，参照 features/market 的 M1 竖切范式逐步完善。"
      />
    </div>
  );
}
