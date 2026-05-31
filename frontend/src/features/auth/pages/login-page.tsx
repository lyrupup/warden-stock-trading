import { useState, type FormEvent } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/auth-store";

type TLocationState = { from?: { pathname?: string } } | null;

/**
 * 登录页 /login（公开）。
 * V1 单用户模式下守卫会自动登录，此页主要为多用户预留 + 兜底入口。
 */
export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [username, setUsername] = useState("");

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    // 脚手架阶段：直接以本地 token 登录（真实登录接入 /auth/login 后替换）
    setAuth("dev-single-user-token", username || "default");
    const state = location.state as TLocationState;
    navigate(state?.from?.pathname ?? "/dashboard", { replace: true });
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-lg">{t("app.fullName")}</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={handleSubmit}>
            <Input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="用户名（脚手架可留空）"
            />
            <Button type="submit" className="w-full">
              {t("common.confirm")}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
