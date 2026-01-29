import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { apiClient } from "../api/client";

/**
 * Login page for API key authentication.
 *
 * Users must enter a valid API key to access the application.
 * The API key is validated against the backend before granting access.
 */
export default function Login() {
  const { login, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [apiKey, setApiKey] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Handle auth token from Cloudflare Access Gate (via cookie, not URL)
  useEffect(() => {
    // Check if already authenticated
    if (isAuthenticated) {
      navigate("/", { replace: true });
      return;
    }

    // Check for session token passed via secure redirect
    const authToken = searchParams.get("civitas_auth");
    if (authToken) {
      // Clear the token from URL immediately to prevent leakage
      window.history.replaceState({}, document.title, window.location.pathname);

      // Validate the session token with the backend
      validateSessionToken(authToken);
    }
  }, [isAuthenticated, navigate, searchParams]);

  const validateSessionToken = async (token: string) => {
    setIsLoading(true);
    setError(null);
    try {
      // Validate the session token with our backend
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

    // Validate API key format (basic check)
    if (trimmedKey.length < 10) {
      setError("Invalid API key format.");
      setIsLoading(false);
      return;
    }

    try {
      // Validate the API key with the backend
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
        setError("Failed to authenticate. Please try again.");
      }
    } finally {
      setIsLoading(false);
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

        <p className="mt-6 text-center text-xs text-blue-300/60">
          Contact your administrator if you need an API key
        </p>
      </div>
    </div>
  );
}
