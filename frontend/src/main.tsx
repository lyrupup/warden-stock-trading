import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "@/core/i18n";
import "@/styles/globals.css";
import { initThemeClass } from "@/stores/theme-store";
import { App } from "./App";

initThemeClass();

async function enableMocking(): Promise<void> {
  if (import.meta.env.VITE_ENABLE_MOCK !== "true") return;
  const { worker } = await import("@/mocks/browser");
  await worker.start({ onUnhandledRequest: "bypass" });
}

void enableMocking().then(() => {
  const root = document.getElementById("root");
  if (!root) throw new Error("Root element #root not found");
  createRoot(root).render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
});
