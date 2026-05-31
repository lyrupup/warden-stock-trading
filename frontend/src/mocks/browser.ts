import { setupWorker } from "msw/browser";
import { handlers } from "./handlers";

/** 浏览器端 MSW worker（开发环境，VITE_ENABLE_MOCK=true 时启用） */
export const worker = setupWorker(...handlers);
