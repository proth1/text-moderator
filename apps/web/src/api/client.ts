import axios from "axios";

const API_BASE_URL = import.meta.env.VITE_API_URL || "/api/v1";

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  // SECURITY: Include cookies for cross-origin requests (required for HttpOnly session cookies)
  withCredentials: true,
});

// Interface for session storage (must match authStore)
interface StoredSession {
  apiKey: string;
  user: unknown;
  expiresAt: number;
}

// Request interceptor to add API key
apiClient.interceptors.request.use(
  (config) => {
    // Try to get API key from session storage
    try {
      const sessionStr = localStorage.getItem("civitas_session");
      if (sessionStr) {
        const session: StoredSession = JSON.parse(sessionStr);
        // SECURITY: Check session expiry before using
        if (session.expiresAt && Date.now() <= session.expiresAt && session.apiKey) {
          config.headers["X-API-Key"] = session.apiKey;
        }
      } else {
        // Legacy support
        const apiKey = localStorage.getItem("api_key");
        if (apiKey) {
          config.headers["X-API-Key"] = apiKey;
        }
      }
    } catch {
      // Ignore errors reading session
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Server responded with error status
      // SECURITY: Don't log full error details in production
      const status = error.response.status;

      // Handle unauthorized - session may have expired
      if (status === 401) {
        // Clear potentially expired session
        localStorage.removeItem("civitas_session");
        localStorage.removeItem("api_key");
        localStorage.removeItem("user");
        console.warn("Session expired or invalid - please log in again");
      }
    } else if (error.request) {
      // Request made but no response
      console.error("Network error - please check your connection");
    }
    return Promise.reject(error);
  }
);

export default apiClient;
