/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_SINGLE_USER_MODE?: string;
  readonly VITE_DEFAULT_TOKEN?: string;
  readonly VITE_SSE_PATH?: string;
  readonly VITE_ENABLE_MOCK?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
