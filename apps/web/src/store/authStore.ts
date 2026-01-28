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

// Read localStorage synchronously on store creation so
// isAuthenticated is correct on the very first render.
function getInitialState() {
  try {
    const apiKey = localStorage.getItem("api_key");
    const userStr = localStorage.getItem("user");
    if (apiKey && userStr) {
      const user = JSON.parse(userStr) as User;
      return { apiKey, user, isAuthenticated: true };
    }
  } catch {
    // Ignore parse errors
  }
  return { apiKey: null, user: null, isAuthenticated: false };
}

const initial = getInitialState();

export const useAuthStore = create<AuthState>((set) => ({
  user: initial.user,
  apiKey: initial.apiKey,
  isAuthenticated: initial.isAuthenticated,

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
