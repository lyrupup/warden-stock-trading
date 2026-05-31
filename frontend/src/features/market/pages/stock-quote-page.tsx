import { useEffect, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { StockQuotePanel } from "../components/stock-quote-panel";
import { MarketStatus } from "../components/market-status";
import { useStockSearch } from "../hooks/use-stock-search";
import type { TStockBrief } from "../types";

/**
 * 个股行情页 /market/quote（M1，左侧「侦查」分区独立入口）。
 * 搜索股票代码 / 名称 → 选中后展示该股当日行情 + 历史 K 线（复用 StockQuotePanel，
 * 与 /market/:code 详情页同一面板，见 FRONTEND.md §6 M1）。
 */
export function StockQuotePage() {
  const { t } = useTranslation();
  const [input, setInput] = useState("");
  const [keyword, setKeyword] = useState(""); // 已提交的搜索词
  const [selected, setSelected] = useState<TStockBrief | null>(null);

  const searchQuery = useStockSearch(keyword);
  const results = searchQuery.data ?? [];

  // 唯一匹配（如精确代码）直接选中，省去一次点击。
  useEffect(() => {
    if (keyword && !selected && results.length === 1) setSelected(results[0]);
  }, [keyword, results, selected]);

  function handleSearch(e: FormEvent) {
    e.preventDefault();
    const v = input.trim();
    if (!v) return;
    setSelected(null);
    setKeyword(v);
  }

  function handlePick(item: TStockBrief) {
    setSelected(item);
    setInput(`${item.stock_name} ${item.stock_code}`);
  }

  return (
    <div>
      <PageHeader
        title={t("market.stockQuote.title")}
        description={t("market.stockQuote.subtitle")}
        actions={<MarketStatus />}
      />

      <form className="mb-4 flex items-center gap-2" onSubmit={handleSearch}>
        <Input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={t("market.search.placeholder")}
          className="h-9 w-72"
        />
        <Button type="submit" size="sm" disabled={searchQuery.isFetching}>
          {t("common.search")}
        </Button>
        {selected ? (
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={() => {
              setSelected(null);
              setKeyword("");
              setInput("");
            }}
          >
            {t("market.search.reset")}
          </Button>
        ) : null}
      </form>

      {/* 搜索结果列表：已搜索、未选中、且多于一条时供选择 */}
      {keyword && !selected ? (
        searchQuery.isError ? (
          <EmptyState
            title={t("common.error")}
            actionLabel={t("common.retry")}
            onAction={() => void searchQuery.refetch()}
          />
        ) : searchQuery.isLoading ? (
          <p className="text-sm text-muted-foreground">{t("common.loading")}</p>
        ) : results.length === 0 ? (
          <EmptyState title={t("market.search.noResult")} />
        ) : (
          <ul className="mb-4 divide-y rounded-md border">
            {results.map((item) => (
              <li key={`${item.market ?? ""}-${item.stock_code}`}>
                <button
                  type="button"
                  className="flex w-full items-center justify-between px-4 py-2 text-left text-sm hover:bg-muted"
                  onClick={() => handlePick(item)}
                >
                  <span className="font-medium">{item.stock_name}</span>
                  <span className="flex items-center gap-2 text-muted-foreground tabular-nums">
                    {item.market ? <Badge variant="secondary">{item.market}</Badge> : null}
                    {item.stock_code}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        )
      ) : null}

      {selected ? (
        <div>
          <div className="mb-4 flex items-center gap-2">
            <h3 className="text-lg font-semibold">{selected.stock_name}</h3>
            <span className="text-muted-foreground tabular-nums">{selected.stock_code}</span>
            {selected.market ? <Badge variant="secondary">{selected.market}</Badge> : null}
          </div>
          <StockQuotePanel code={selected.stock_code} />
        </div>
      ) : !keyword ? (
        <EmptyState title={t("market.search.hint")} />
      ) : null}
    </div>
  );
}
