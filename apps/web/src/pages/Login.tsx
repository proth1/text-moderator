import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { UserRole } from "../types";

/**
 * Login page that auto-authenticates users.
 *
 * Since users have already passed through the Cloudflare Access Gate
 * (password + magic link), we auto-login with the default admin credentials.
 * This provides a seamless experience after external authentication.
 */
export default function Login() {
  const { login, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const hasAutoLoggedIn = useRef(false);

  // User mapping for valid API keys
  const usersByApiKey: Record<string, { id: string; email: string; role: UserRole }> = {
    "tk_admin_test_key_001": { id: "a0000000-0000-0000-0000-000000000001", email: "admin@civitas.test", role: UserRole.ADMIN },
    "tk_mod_test_key_002": { id: "a0000000-0000-0000-0000-000000000002", email: "moderator@civitas.test", role: UserRole.MODERATOR },
    "tk_viewer_test_key_003": { id: "a0000000-0000-0000-0000-000000000003", email: "viewer@civitas.test", role: UserRole.VIEWER },
  };

  useEffect(() => {
    // Already authenticated - redirect to dashboard
    if (isAuthenticated) {
      navigate("/", { replace: true });
      return;
    }

    // Auto-login once with admin credentials
    if (!hasAutoLoggedIn.current) {
      hasAutoLoggedIn.current = true;

      const apiKey = "tk_admin_test_key_001";
      const userInfo = usersByApiKey[apiKey];

      const mockUser = {
        ...userInfo,
        created_at: new Date().toISOString(),
      };

      login(apiKey, mockUser);
      navigate("/", { replace: true });
    }
  }, [isAuthenticated, login, navigate]);

  // Show elegant loading state while auto-login happens
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-blue-900 to-slate-900">
      <div className="text-center">
        {/* Animated logo */}
        <div className="relative mb-8">
          <div className="w-20 h-20 mx-auto bg-blue-500 rounded-2xl flex items-center justify-center shadow-2xl shadow-blue-500/30 animate-pulse">
            <svg className="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          {/* Animated ring */}
          <div className="absolute inset-0 w-20 h-20 mx-auto rounded-2xl border-2 border-blue-400/30 animate-ping" style={{ animationDuration: '2s' }} />
        </div>

        <h1 className="text-3xl font-bold text-white mb-2">Civitas AI</h1>
        <p className="text-blue-200 mb-8">Trust & Safety Infrastructure</p>

        {/* Loading indicator */}
        <div className="flex items-center justify-center gap-2 text-blue-300">
          <svg className="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <span className="text-sm">Authenticating...</span>
        </div>
      </div>
    </div>
  );
}
