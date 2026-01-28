import { create } from "zustand";
import type { User } from "../types";

interface AuthState {
  user: User | null;
  apiKey: string | null;
  isAuthenticated: boolean;
  login: (apiKey: string, user: User) => void;
  logout: () => void;
  checkAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  apiKey: null,
  isAuthenticated: false,

  login: (apiKey: string, user: User) => {
    localStorage.setItem("api_key", apiKey);
    localStorage.setItem("user", JSON.stringify(user));
    set({ apiKey, user, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem("api_key");
    localStorage.removeItem("user");
    set({ apiKey: null, user: null, isAuthenticated: false });
  },

  checkAuth: () => {
    const apiKey = localStorage.getItem("api_key");
    const userStr = localStorage.getItem("user");

    if (apiKey && userStr) {
      try {
        const user = JSON.parse(userStr);
        set({ apiKey, user, isAuthenticated: true });
      } catch (e) {
        localStorage.removeItem("api_key");
        localStorage.removeItem("user");
        set({ apiKey: null, user: null, isAuthenticated: false });
      }
    }
  },
}));
