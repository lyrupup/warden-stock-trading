import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MarketStatus } from "../components/market-status";
import { IndexCard } from "../components/index-card";
import { WatchlistTable } from "../components/watchlist-table";
import { useIndices } from "../hooks/use-indices";
import { useWatchlistQuotes } from "../hooks/use-watchlist-quotes";
import { useAddWatch, useRemoveWatch, useWatchlist } from "../hooks/use-watchlist";

/** 行情中心 /market（M1 竖切范式） */
export function MarketPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [newCode, setNewCode] = useState("");

  const indicesQuery = useIndices();
  const quotesQuery = useWatchlistQuotes();
  const watchlistQuery = useWatchlist();
  const addWatch = useAddWatch();
  const removeWatch = useRemoveWatch();

  const quotes = quotesQuery.data ?? [];
  const hasStale = quotes.some((q) => q.stale);

  function handleAdd(e: FormEvent) {
    e.preventDefault();
    const code = newCode.trim();
    if (!code) return;
    addWatch.mutate({ stock_code: code }, { onSuccess: () => setNewCode("") });
  }

  function handleRemove(stockCode: string) {
    const item = watchlistQuery.data?.find((w) => w.stock_code === stockCode);
    if (item) removeWatch.mutate(item.id);
  }

  return (
    <div>
      <PageHeader title={t("market.title")} description={t("market.subtitle")} actions={<MarketStatus />} />

      {hasStale ? (
        <div className="mb-4 rounded-md border border-quote-flat/30 bg-muted px-4 py-2 text-sm text-muted-foreground">
          {t("market.staleTip")}
        </div>
      ) : null}

      <section className="mb-8">
        <h2 className="mb-3 text-sm font-medium text-muted-foreground">{t("market.indices")}</h2>
        {indicesQuery.isError ? (
          <EmptyState
            title={t("common.error")}
            actionLabel={t("common.retry")}
            onAction={() => void indicesQuery.refetch()}
          />
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
            {(indicesQuery.data ?? []).map((index) => (
              <IndexCard key={index.index_code} quote={index} />
            ))}
          </div>
        )}
      </section>

      <section>
        <div className="mb-3 flex items-center justify-between gap-4">
          <h2 className="text-sm font-medium text-muted-foreground">{t("market.watchlist")}</h2>
          <form className="flex items-center gap-2" onSubmit={handleAdd}>
            <Input
              value={newCode}
              onChange={(e) => setNewCode(e.target.value)}
              placeholder={t("market.stockCode")}
              className="h-8 w-40"
            />
            <Button type="submit" size="sm" disabled={addWatch.isPending}>
              {t("market.addWatch")}
            </Button>
          </form>
        </div>
        {quotesQuery.isError ? (
          <EmptyState
            title={t("common.error")}
            actionLabel={t("common.retry")}
            onAction={() => void quotesQuery.refetch()}
          />
        ) : (
          <WatchlistTable
            data={quotes}
            loading={quotesQuery.isLoading}
            onRowClick={(q) => navigate(`/market/${q.stock_code}`)}
            onRemove={handleRemove}
          />
        )}
      </section>
    </div>
  );
}
