import { useAuthStore } from "@/stores/auth-store";

/** 是否开启单用户模式（V1） */
export function isSingleUserMode(): boolean {
  return import.meta.env.VITE_SINGLE_USER_MODE === "true";
}

/**
 * 单用户模式自动登录：若开启且当前无 token，则注入默认 token。
 * 见 FRONTEND.md §5.1 多用户预留。
 */
export function ensureSingleUserAuth(): void {
  if (!isSingleUserMode()) return;
  const { token, setAuth } = useAuthStore.getState();
  if (!token) {
    const defaultToken = import.meta.env.VITE_DEFAULT_TOKEN ?? "dev-single-user-token";
    setAuth(defaultToken, "default");
  }
}
