import { type ReactNode } from "react";
import { cn } from "@/lib/cn";
import { Button } from "@/components/ui/button";

type TEmptyStateProps = {
  title?: string;
  description?: string;
  icon?: ReactNode;
  /** 操作（如重试） */
  actionLabel?: string;
  onAction?: () => void;
  className?: string;
};

/** 空态 / 错误降级（见 FRONTEND.md §2 common/empty-state） */
export function EmptyState({
  title = "暂无数据",
  description,
  icon,
  actionLabel,
  onAction,
  className,
}: TEmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed p-10 text-center",
        className,
      )}
    >
      {icon ? <div className="text-muted-foreground">{icon}</div> : null}
      <div className="space-y-1">
        <p className="text-sm font-medium">{title}</p>
        {description ? <p className="text-xs text-muted-foreground">{description}</p> : null}
      </div>
      {actionLabel && onAction ? (
        <Button variant="outline" size="sm" onClick={onAction}>
          {actionLabel}
        </Button>
      ) : null}
    </div>
  );
}
