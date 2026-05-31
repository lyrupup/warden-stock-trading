import ky from "ky";
import { useAuthStore } from "@/stores/auth-store";
import { AppError } from "./app-error";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api";

/**
 * 统一 ky 实例（见 FRONTEND.md §4.1）：
 * - beforeRequest：注入 Bearer token
 * - afterResponse：401 登出；成功响应校验业务 code，非 0 抛 AppError
 */
export const httpClient = ky.create({
  prefixUrl: API_BASE_URL,
  timeout: 30_000,
  retry: 0,
  hooks: {
    beforeRequest: [
      (request) => {
        const token = useAuthStore.getState().token;
        if (token) request.headers.set("Authorization", `Bearer ${token}`);
      },
    ],
    afterResponse: [
      async (_request, _options, response) => {
        if (response.status === 401) {
          useAuthStore.getState().logout();
          return;
        }
        if (response.ok) {
          const body = (await response
            .clone()
            .json()
            .catch(() => null)) as { code: number; message: string } | null;
          if (body && body.code !== 0) {
            throw new AppError(body.code, body.message);
          }
        }
      },
    ],
  },
});
