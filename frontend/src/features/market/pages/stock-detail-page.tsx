import { useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { PageHeader } from "@/components/common/page-header";
import { StockQuotePanel } from "../components/stock-quote-panel";
import { useStockQuote } from "../hooks/use-stock-detail";

/** 个股详情 /market/:code（M1） */
export function StockDetailPage() {
  const { code = "" } = useParams();
  const { t } = useTranslation();

  const quoteQuery = useStockQuote(code);
  const quote = quoteQuery.data;

  return (
    <div>
      <PageHeader
        title={quote ? `${quote.stock_name} ${quote.stock_code}` : code}
        description={t("market.title")}
      />
      <StockQuotePanel code={code} />
    </div>
  );
}
