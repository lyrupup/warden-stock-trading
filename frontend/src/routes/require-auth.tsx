import { useEffect, useState, type ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useAuthStore } from "@/stores/auth-store";
import { ensureSingleUserAuth, isSingleUserMode } from "@/core/auth";

type TRequireAuthProps = {
  children: ReactNode;
};

/**
 * 登录守卫：读取 auth-store.token。
 * 单用户模式（VITE_SINGLE_USER_MODE=true）下自动注入默认 token，不跳转登录。
 */
export function RequireAuth({ children }: TRequireAuthProps) {
  const location = useLocation();
  const token = useAuthStore((s) => s.token);
  const [ready, setReady] = useState(!isSingleUserMode());

  useEffect(() => {
    if (isSingleUserMode()) {
      ensureSingleUserAuth();
      setReady(true);
    }
  }, []);

  if (!ready) return null;

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return <>{children}</>;
}
