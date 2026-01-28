import { useAuthStore } from "../store/authStore";

export function useAuth() {
  const { user, apiKey, isAuthenticated, login, logout, checkAuth } = useAuthStore();

  return {
    user,
    apiKey,
    isAuthenticated,
    login,
    logout,
    checkAuth,
  };
}
