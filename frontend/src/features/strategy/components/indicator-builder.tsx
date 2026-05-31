import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/cn";
import type {
  TFactorExpr,
  TFactorType,
  TIndicatorCatalogItem,
  TIndicatorGroup,
  TIndicatorRule,
} from "../types";
import { OPS, applyParam, defaultFactor, isBoolFactor } from "./factor-utils";

type TIndicatorBuilderProps = {
  value: TIndicatorGroup;
  catalog: TIndicatorCatalogItem[];
  onChange: (group: TIndicatorGroup) => void;
};

const selectClass =
  "h-9 rounded-md border border-input bg-background px-2 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring";

/**
 * 量化指标构造器（FRONTEND.md §M2）：可视化增删规则行，每行「左因子 op 右值」，
 * 因子可选项与参数由因子目录驱动，支持 and/or。V1 仅编辑顶层规则（嵌套子组由后端 schema 支持）。
 */
export function IndicatorBuilder({ value, catalog, onChange }: TIndicatorBuilderProps) {
  function setLogic(logic: "and" | "or") {
    onChange({ ...value, logic });
  }

  function updateRule(index: number, rule: TIndicatorRule) {
    const rules = value.rules.slice();
    rules[index] = rule;
    onChange({ ...value, rules });
  }

  function addRule() {
    const left = defaultFactor("ma", catalog);
    const right = defaultFactor("ma", catalog);
    onChange({ ...value, rules: [...value.rules, { left, op: ">", right }] });
  }

  function removeRule(index: number) {
    onChange({ ...value, rules: value.rules.filter((_, i) => i !== index) });
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 text-sm">
        <span className="text-muted-foreground">满足以下</span>
        <select
          className={selectClass}
          value={value.logic}
          onChange={(e) => setLogic(e.target.value as "and" | "or")}
        >
          <option value="and">全部（与）</option>
          <option value="or">任一（或）</option>
        </select>
        <span className="text-muted-foreground">条件：</span>
      </div>

      <div className="space-y-3">
        {value.rules.map((rule, i) => (
          <RuleEditor
            key={i}
            rule={rule}
            catalog={catalog}
            onChange={(r) => updateRule(i, r)}
            onRemove={() => removeRule(i)}
          />
        ))}
        {value.rules.length === 0 ? (
          <p className="rounded-md border border-dashed px-3 py-4 text-center text-sm text-muted-foreground">
            暂无规则，点击下方「添加规则」开始构建量化条件
          </p>
        ) : null}
      </div>

      <Button type="button" variant="outline" size="sm" onClick={addRule}>
        + 添加规则
      </Button>
    </div>
  );
}

type TRuleEditorProps = {
  rule: TIndicatorRule;
  catalog: TIndicatorCatalogItem[];
  onChange: (rule: TIndicatorRule) => void;
  onRemove: () => void;
};

function RuleEditor({ rule, catalog, onChange, onRemove }: TRuleEditorProps) {
  const leftIsBool = isBoolFactor(rule.left.type);

  function setLeft(left: TFactorExpr) {
    // 左因子切换为布尔类时，自动把右值规整为布尔常量并用 == 比较。
    if (isBoolFactor(left.type)) {
      onChange({ left, op: "==", right: { type: "const", value: true } });
      return;
    }
    onChange({ ...rule, left });
  }

  return (
    <div className="flex flex-wrap items-center gap-2 rounded-md border bg-card p-3">
      <FactorEditor expr={rule.left} catalog={catalog} allowConst={false} onChange={setLeft} />

      <select
        className={selectClass}
        value={rule.op}
        onChange={(e) => onChange({ ...rule, op: e.target.value as TIndicatorRule["op"] })}
        disabled={leftIsBool}
      >
        {OPS.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>

      <FactorEditor
        expr={rule.right}
        catalog={catalog}
        allowConst
        boolOnly={leftIsBool}
        onChange={(right) => onChange({ ...rule, right })}
      />

      <Button type="button" variant="ghost" size="sm" className="ml-auto" onClick={onRemove}>
        删除
      </Button>
    </div>
  );
}

type TFactorEditorProps = {
  expr: TFactorExpr;
  catalog: TIndicatorCatalogItem[];
  allowConst: boolean;
  boolOnly?: boolean;
  onChange: (expr: TFactorExpr) => void;
};

function FactorEditor({ expr, catalog, allowConst, boolOnly, onChange }: TFactorEditorProps) {
  // 布尔规则右值固定为真/假常量。
  if (boolOnly) {
    return (
      <select
        className={selectClass}
        value={expr.value === true ? "true" : "false"}
        onChange={(e) => onChange({ type: "const", value: e.target.value === "true" })}
      >
        <option value="true">成立（真）</option>
        <option value="false">不成立（假）</option>
      </select>
    );
  }

  const options: { type: TFactorType; name: string }[] = catalog.map((c) => ({ type: c.type, name: c.name }));
  if (allowConst) options.push({ type: "const", name: "常量" });

  function setType(type: TFactorType) {
    onChange(defaultFactor(type, catalog));
  }

  const item = catalog.find((c) => c.type === expr.type);

  return (
    <div className="flex flex-wrap items-center gap-1.5">
      <select className={selectClass} value={expr.type} onChange={(e) => setType(e.target.value as TFactorType)}>
        {options.map((o) => (
          <option key={o.type} value={o.type}>
            {o.name}
          </option>
        ))}
      </select>

      {expr.type === "const" ? (
        <ConstInput expr={expr} onChange={onChange} />
      ) : (
        item?.params.map((p) => (
          <ParamInput key={p.key} paramKey={p.key} spec={p} expr={expr} onChange={onChange} />
        ))
      )}
    </div>
  );
}

function ConstInput({ expr, onChange }: { expr: TFactorExpr; onChange: (e: TFactorExpr) => void }) {
  const raw = typeof expr.value === "boolean" ? String(expr.value) : (expr.value ?? 0);
  return (
    <Input
      className="h-9 w-24"
      value={String(raw)}
      onChange={(e) => {
        const v = e.target.value.trim();
        if (v === "true") return onChange({ type: "const", value: true });
        if (v === "false") return onChange({ type: "const", value: false });
        const n = Number(v);
        onChange({ type: "const", value: Number.isFinite(n) ? n : 0 });
      }}
    />
  );
}

type TParamInputProps = {
  paramKey: string;
  spec: TIndicatorCatalogItem["params"][number];
  expr: TFactorExpr;
  onChange: (e: TFactorExpr) => void;
};

function ParamInput({ paramKey, spec, expr, onChange }: TParamInputProps) {
  function update(value: unknown) {
    const next = { ...expr };
    applyParam(next, paramKey, value);
    onChange(next);
  }

  if (spec.type === "enum") {
    const current = paramKey === "direction" ? (expr.direction ?? "bull") : (expr.field ?? "");
    return (
      <select className={selectClass} value={current} onChange={(e) => update(e.target.value)} title={spec.desc}>
        {(spec.enum ?? []).map((opt) => (
          <option key={String(opt)} value={String(opt)}>
            {String(opt)}
          </option>
        ))}
      </select>
    );
  }

  if (spec.type === "int[]") {
    const current = (expr.periods ?? []).join(",");
    return (
      <Input
        className={cn("h-9 w-28")}
        value={current}
        title={spec.desc}
        placeholder="5,10,20"
        onChange={(e) =>
          update(
            e.target.value
              .split(",")
              .map((s) => Number(s.trim()))
              .filter((n) => Number.isFinite(n) && n > 0),
          )
        }
      />
    );
  }

  // int / float
  const numVal =
    paramKey === "period"
      ? expr.period
      : paramKey === "threshold"
        ? expr.threshold
        : paramKey === "days"
          ? expr.days
          : paramKey === "eps"
            ? expr.eps
            : undefined;
  return (
    <Input
      className="h-9 w-20"
      type="number"
      value={numVal ?? ""}
      title={spec.desc}
      onChange={(e) => update(e.target.value)}
    />
  );
}
