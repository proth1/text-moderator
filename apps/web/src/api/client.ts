import axios from "axios";

const API_BASE_URL = import.meta.env.VITE_API_URL || "/api/v1";

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor to add API key
apiClient.interceptors.request.use(
  (config) => {
    const apiKey = localStorage.getItem("api_key");
    if (apiKey) {
      config.headers["X-API-Key"] = apiKey;
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
      const message = error.response.data?.error || error.response.data?.message || "An error occurred";
      console.error("API Error:", message);

      // Handle unauthorized
      if (error.response.status === 401) {
        localStorage.removeItem("api_key");
        window.location.href = "/login";
      }
    } else if (error.request) {
      // Request made but no response
      console.error("Network Error:", error.message);
    } else {
      console.error("Error:", error.message);
    }
    return Promise.reject(error);
  }
);

export default apiClient;
