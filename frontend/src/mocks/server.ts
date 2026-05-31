import { setupServer } from "msw/node";
import { handlers } from "./handlers";

/** Node 端 MSW server（Vitest 测试使用） */
export const server = setupServer(...handlers);
