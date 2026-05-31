import { create } from "zustand";

const TOKEN_KEY = "warden_token";
const USER_ID_KEY = "warden_user_id";

/** 认证状态（多用户预留，见 FRONTEND.md §4.4） */
export type TAuthState = {
  token: string | null;
  userId: string | null;
  isAuthenticated: () => boolean;
  setAuth: (token: string, userId: string) => void;
  logout: () => void;
};

export const useAuthStore = create<TAuthState>((set, get) => ({
  token: localStorage.getItem(TOKEN_KEY),
  userId: localStorage.getItem(USER_ID_KEY),
  isAuthenticated: () => Boolean(get().token),
  setAuth: (token, userId) => {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_ID_KEY, userId);
    set({ token, userId });
  },
  logout: () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_ID_KEY);
    set({ token: null, userId: null });
  },
}));
