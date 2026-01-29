import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { apiClient } from "../api/client";
import { UserRole } from "../types";

/**
 * Login page for API key authentication.
 *
 * In development mode, predefined test API keys are available.
 * In production, API keys must be validated against the backend.
 */

// Development mode test users (only available when VITE_DEV_MODE=true or no API configured)
const DEV_USERS: Record<string, { id: string; email: string; role: UserRole }> = {
  "tk_admin_test_key_001": { id: "a0000000-0000-0000-0000-000000000001", email: "admin@civitas.dev", role: UserRole.ADMIN },
  "tk_mod_test_key_002": { id: "a0000000-0000-0000-0000-000000000002", email: "moderator@civitas.dev", role: UserRole.MODERATOR },
  "tk_viewer_test_key_003": { id: "a0000000-0000-0000-0000-000000000003", email: "viewer@civitas.dev", role: UserRole.VIEWER },
};

const isDevelopment = import.meta.env.DEV || import.meta.env.VITE_DEV_MODE === "true";

export default function Login() {
  const { login, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [apiKey, setApiKey] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      navigate("/", { replace: true });
      return;
    }

    // Handle auth token from Cloudflare Access Gate
    const authToken = searchParams.get("civitas_auth");
    if (authToken) {
      window.history.replaceState({}, document.title, window.location.pathname);
      validateSessionToken(authToken);
    }
  }, [isAuthenticated, navigate, searchParams]);

  const validateSessionToken = async (token: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiClient.post("/auth/validate-session", { token });
      if (response.data?.valid && response.data?.user) {
        login(response.data.apiKey, response.data.user);
        navigate("/", { replace: true });
      } else {
        setError("Invalid session. Please log in again.");
      }
    } catch {
      setError("Failed to validate session. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    const trimmedKey = apiKey.trim();
    if (!trimmedKey) {
      setError("Please enter an API key.");
      setIsLoading(false);
      return;
    }

    // Development mode: check against test credentials
    if (isDevelopment && DEV_USERS[trimmedKey]) {
      const devUser = DEV_USERS[trimmedKey];
      login(trimmedKey, {
        ...devUser,
        created_at: new Date().toISOString(),
      });
      navigate("/", { replace: true });
      setIsLoading(false);
      return;
    }

    // Production mode: validate with backend
    try {
      const response = await apiClient.get("/auth/validate", {
        headers: { "X-API-Key": trimmedKey },
      });

      if (response.data?.user) {
        login(trimmedKey, response.data.user);
        navigate("/", { replace: true });
      } else {
        setError("Invalid API key.");
      }
    } catch (err: unknown) {
      const errorResponse = err as { response?: { status?: number } };
      if (errorResponse.response?.status === 401) {
        setError("Invalid API key. Please check and try again.");
      } else {
        // In development, show hint about test keys
        if (isDevelopment) {
          setError("API unavailable. Use a test key: tk_admin_test_key_001");
        } else {
          setError("Failed to authenticate. Please try again.");
        }
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleQuickLogin = (key: string) => {
    setApiKey(key);
    const devUser = DEV_USERS[key];
    if (devUser) {
      login(key, {
        ...devUser,
        created_at: new Date().toISOString(),
      });
      navigate("/", { replace: true });
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-blue-900 to-slate-900">
      <div className="w-full max-w-md p-8">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="relative mb-6">
            <div className="w-20 h-20 mx-auto bg-blue-500 rounded-2xl flex items-center justify-center shadow-2xl shadow-blue-500/30">
              <svg className="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
          </div>
          <h1 className="text-3xl font-bold text-white mb-2">Civitas AI</h1>
          <p className="text-blue-200">Trust & Safety Infrastructure</p>
        </div>

        {/* Login Form */}
        <form onSubmit={handleSubmit} className="bg-white/10 backdrop-blur-lg rounded-2xl p-6 border border-white/20">
          <div className="mb-6">
            <label htmlFor="apiKey" className="block text-sm font-medium text-blue-100 mb-2">
              API Key
            </label>
            <input
              type="password"
              id="apiKey"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="Enter your API key"
              autoComplete="current-password"
              disabled={isLoading}
              className="w-full px-4 py-3 bg-white/10 border border-white/20 rounded-lg text-white placeholder-blue-300/50 focus:ring-2 focus:ring-blue-400 focus:border-transparent transition disabled:opacity-50"
            />
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-500/20 border border-red-500/50 rounded-lg text-red-200 text-sm">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={isLoading}
            className="w-full bg-blue-500 hover:bg-blue-600 disabled:bg-blue-500/50 text-white py-3 px-4 rounded-lg font-semibold transition duration-200 flex items-center justify-center gap-2"
          >
            {isLoading ? (
              <>
                <svg className="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                <span>Authenticating...</span>
              </>
            ) : (
              "Sign In"
            )}
          </button>
        </form>

        {/* Development Quick Login */}
        {isDevelopment && (
          <div className="mt-6 p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl">
            <p className="text-amber-200 text-xs font-medium mb-3 flex items-center gap-2">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              Development Mode - Quick Login
            </p>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => handleQuickLogin("tk_admin_test_key_001")}
                className="px-3 py-1.5 bg-purple-500/20 hover:bg-purple-500/30 border border-purple-500/40 rounded-lg text-purple-200 text-xs font-medium transition"
              >
                Admin
              </button>
              <button
                type="button"
                onClick={() => handleQuickLogin("tk_mod_test_key_002")}
                className="px-3 py-1.5 bg-blue-500/20 hover:bg-blue-500/30 border border-blue-500/40 rounded-lg text-blue-200 text-xs font-medium transition"
              >
                Moderator
              </button>
              <button
                type="button"
                onClick={() => handleQuickLogin("tk_viewer_test_key_003")}
                className="px-3 py-1.5 bg-slate-500/20 hover:bg-slate-500/30 border border-slate-500/40 rounded-lg text-slate-200 text-xs font-medium transition"
              >
                Viewer
              </button>
            </div>
          </div>
        )}

        <p className="mt-6 text-center text-xs text-blue-300/60">
          {isDevelopment
            ? "Development mode - test credentials available above"
            : "Contact your administrator if you need an API key"
          }
        </p>
      </div>
    </div>
  );
}
