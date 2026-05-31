import { type ReactNode } from "react";
import { cn } from "@/lib/cn";

type TPageHeaderProps = {
  title: string;
  description?: string;
  /** 右侧操作区 */
  actions?: ReactNode;
  className?: string;
};

/** 页面标题 + 操作区（见 FRONTEND.md §5.2） */
export function PageHeader({ title, description, actions, className }: TPageHeaderProps) {
  return (
    <div className={cn("mb-6 flex items-start justify-between gap-4", className)}>
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {description ? <p className="text-sm text-muted-foreground">{description}</p> : null}
      </div>
      {actions ? <div className="flex items-center gap-2">{actions}</div> : null}
    </div>
  );
}
