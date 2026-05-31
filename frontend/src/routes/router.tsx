import { createBrowserRouter, Navigate } from "react-router-dom";
import { LoginPage } from "@/features/auth";
import { DashboardPage } from "@/features/dashboard";
import { MarketPage, StockDetailPage } from "@/features/market";
import { StrategyDetailPage, StrategyListPage } from "@/features/strategy";
import { PositionDetailPage, PositionListPage } from "@/features/position";
import { PremarketPage, RiskPage } from "@/features/risk";
import { TaskListPage } from "@/features/task";
import { AiAnalysisPage, AiReportsPage } from "@/features/ai-analysis";
import { SettingsPage } from "@/features/settings";
import { AppLayout } from "./app-layout";
import { RequireAuth } from "./require-auth";

export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    path: "/",
    element: (
      <RequireAuth>
        <AppLayout />
      </RequireAuth>
    ),
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: "dashboard", element: <DashboardPage /> },
      { path: "market", element: <MarketPage /> },
      { path: "market/:code", element: <StockDetailPage /> },
      { path: "strategies", element: <StrategyListPage /> },
      { path: "strategies/:id", element: <StrategyDetailPage /> },
      { path: "positions", element: <PositionListPage /> },
      { path: "positions/:id", element: <PositionDetailPage /> },
      { path: "risk", element: <RiskPage /> },
      { path: "risk/premarket", element: <PremarketPage /> },
      { path: "tasks", element: <TaskListPage /> },
      { path: "ai", element: <AiAnalysisPage /> },
      { path: "ai/reports", element: <AiReportsPage /> },
      { path: "settings", element: <SettingsPage /> },
    ],
  },
  { path: "*", element: <Navigate to="/dashboard" replace /> },
]);
