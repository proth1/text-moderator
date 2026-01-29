import { create } from "zustand";
import type { User } from "../types";

/**
 * SECURITY NOTICE:
 * This implementation stores API keys in localStorage which is vulnerable to XSS attacks.
 * For production, consider:
 * 1. Using HttpOnly cookies set by the server (implemented in access-gate)
 * 2. Implementing token refresh with short-lived access tokens
 * 3. Adding Content Security Policy headers
 *
 * The current implementation includes:
 * - Session expiry (24 hours)
 * - Session validation on app load
 */

const SESSION_DURATION_MS = 24 * 60 * 60 * 1000; // 24 hours

interface StoredSession {
  apiKey: string;
  user: User;
  expiresAt: number;
}

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
function getInitialState(): { apiKey: string | null; user: User | null; isAuthenticated: boolean } {
  try {
    const sessionStr = localStorage.getItem("civitas_session");
    if (sessionStr) {
      const session: StoredSession = JSON.parse(sessionStr);

      // SECURITY: Check if session has expired
      if (session.expiresAt && Date.now() > session.expiresAt) {
        // Session expired - clean up
        localStorage.removeItem("civitas_session");
        localStorage.removeItem("api_key"); // Clean up legacy storage
        localStorage.removeItem("user");
        return { apiKey: null, user: null, isAuthenticated: false };
      }

      return { apiKey: session.apiKey, user: session.user, isAuthenticated: true };
    }

    // Legacy support: migrate old storage format
    const apiKey = localStorage.getItem("api_key");
    const userStr = localStorage.getItem("user");
    if (apiKey && userStr) {
      const user = JSON.parse(userStr) as User;
      // Migrate to new format with expiry
      const newSession: StoredSession = {
        apiKey,
        user,
        expiresAt: Date.now() + SESSION_DURATION_MS,
      };
      localStorage.setItem("civitas_session", JSON.stringify(newSession));
      // Clean up old format
      localStorage.removeItem("api_key");
      localStorage.removeItem("user");
      return { apiKey, user, isAuthenticated: true };
    }
  } catch {
    // Ignore parse errors - clear potentially corrupted data
    localStorage.removeItem("civitas_session");
    localStorage.removeItem("api_key");
    localStorage.removeItem("user");
  }
  return { apiKey: null, user: null, isAuthenticated: false };
}

const initial = getInitialState();

export const useAuthStore = create<AuthState>((set) => ({
  user: initial.user,
  apiKey: initial.apiKey,
  isAuthenticated: initial.isAuthenticated,

  login: (apiKey: string, user: User) => {
    // Store session with expiry time
    const session: StoredSession = {
      apiKey,
      user,
      expiresAt: Date.now() + SESSION_DURATION_MS,
    };
    localStorage.setItem("civitas_session", JSON.stringify(session));
    set({ apiKey, user, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem("civitas_session");
    // Clean up legacy storage if present
    localStorage.removeItem("api_key");
    localStorage.removeItem("user");
    set({ apiKey: null, user: null, isAuthenticated: false });
  },

  checkAuth: () => {
    try {
      const sessionStr = localStorage.getItem("civitas_session");
      if (sessionStr) {
        const session: StoredSession = JSON.parse(sessionStr);

        // SECURITY: Check if session has expired
        if (session.expiresAt && Date.now() > session.expiresAt) {
          localStorage.removeItem("civitas_session");
          set({ apiKey: null, user: null, isAuthenticated: false });
          return;
        }

        set({ apiKey: session.apiKey, user: session.user, isAuthenticated: true });
        return;
      }
    } catch {
      // Clear corrupted data
      localStorage.removeItem("civitas_session");
    }
    set({ apiKey: null, user: null, isAuthenticated: false });
  },
}));
