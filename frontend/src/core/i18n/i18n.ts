import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import zhCN from "./locales/zh-CN.json";

export const defaultNS = "translation";

export const resources = {
  "zh-CN": { translation: zhCN },
} as const;

void i18n.use(initReactI18next).init({
  resources,
  lng: "zh-CN",
  fallbackLng: "zh-CN",
  defaultNS,
  interpolation: {
    escapeValue: false,
  },
});

export { i18n };
